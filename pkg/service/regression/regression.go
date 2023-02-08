package regression

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-test/deep"
	"github.com/wI2L/jsondiff"

	"github.com/google/uuid"
	"github.com/k0kubun/pp/v3"
	grpcMock "go.keploy.io/server/grpc/mock"
	"go.keploy.io/server/grpc/utils"
	"go.keploy.io/server/pkg"
	"go.keploy.io/server/pkg/models"
	"go.keploy.io/server/pkg/platform/telemetry"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func New(tdb models.TestCaseDB, rdb TestRunDB, testReportFS TestReportFS, adb telemetry.Service, cl http.Client, log *zap.Logger, TestExport bool, mFS models.MockFS) *Regression {
	return &Regression{
		yamlTcs:      sync.Map{},
		tele:         adb,
		tdb:          tdb,
		log:          log,
		client:       cl,
		rdb:          rdb,
		testReportFS: testReportFS,
		mockFS:       mFS,
		testExport:   TestExport,
	}
}

type Regression struct {
	yamlTcs  sync.Map
	runCount int

	tdb          models.TestCaseDB
	client       http.Client
	testReportFS TestReportFS
	rdb          TestRunDB
	tele         telemetry.Service
	mockFS       models.MockFS
	testExport   bool
	log          *zap.Logger
}

func (r *Regression) startTestRun(ctx context.Context, runId, testCasePath, mockPath, testReportPath string, totalTcs int) error {
	if !pkg.IsValidPath(testCasePath) || !pkg.IsValidPath(mockPath) {
		return pkg.LogError("file path should be absolute to read and write testcases and their mocks", r.log, fmt.Errorf("file path should be absolute"))
	}
	tcs, err := r.mockFS.ReadAll(ctx, testCasePath, mockPath)
	if err != nil {
		return pkg.LogError("failed to read and cache testcases from ", r.log, err, map[string]interface{}{"mock path": mockPath, "testcase path": testCasePath})
	}
	tcsMap := sync.Map{}
	for _, j := range tcs {
		tcsMap.Store(j.ID, j)
	}
	r.yamlTcs.Store(runId, tcsMap)
	err = r.testReportFS.Write(ctx, testReportPath, models.TestReport{
		Version: models.V1Beta1,
		Name:    runId,
		Total:   len(tcs),
		Status:  string(models.TestRunStatusRunning),
	})
	if err != nil {
		return pkg.LogError("failed to create test report file", r.log, err, map[string]interface{}{"file path": testReportPath})
	}
	return nil
}

func (r *Regression) stopTestRun(ctx context.Context, runId, testReportPath string) error {
	r.yamlTcs.Delete(runId)
	testResults, err := r.testReportFS.GetResults(runId)
	if err != nil {
		return pkg.LogError("failed to stopTestRun", r.log, err)
	}
	var (
		success = 0
		failure = 0
		status  = models.TestRunStatusPassed
	)
	for _, j := range testResults {
		if j.Status == models.TestStatusPassed {
			success++
		} else if j.Status == models.TestStatusFailed {
			failure++
			status = models.TestRunStatusFailed
		}
	}
	err = r.testReportFS.Write(ctx, testReportPath, models.TestReport{
		Version: models.V1Beta1,
		Name:    runId,
		Total:   len(testResults),
		Status:  string(status),
		Tests:   testResults,
		Success: success,
		Failure: failure,
	})
	if err != nil {
		return pkg.LogError("failed to create test report file", r.log, err, map[string]interface{}{"file path": testReportPath})
	}
	return nil
}

func (r *Regression) test(ctx context.Context, cid, runId, id, app string, resp models.HttpResp) (bool, *models.Result, *models.TestCase, error) {
	var (
		tc  models.TestCase
		err error
	)
	switch r.testExport {
	case false:
		tc, err = r.tdb.Get(ctx, cid, id)
		if err != nil {
			return false, nil, nil, pkg.LogError("failed to get testcase from DB", r.log, err, map[string]interface{}{"id": id, "cid": cid, "appID": app})
		}
	case true:
		if val, ok := r.yamlTcs.Load(runId); ok {
			tcsMap := val.(sync.Map)
			if val, ok := tcsMap.Load(id); ok {
				tc = val.(models.TestCase)
				// tcsMap.Delete(id)
			} else {
				return false, nil, nil, pkg.LogError("", r.log, fmt.Errorf("failed to load testcase from tcs map coresponding to testcaseId: %s", pkg.SanitiseInput(id)))
			}
		} else {
			return false, nil, nil, pkg.LogError("", r.log, fmt.Errorf("failed to load testcases coresponding to runId: %s", pkg.SanitiseInput(runId)))
		}
	}
	bodyType := models.BodyTypePlain
	if json.Valid([]byte(resp.Body)) {
		bodyType = models.BodyTypeJSON
	}
	pass := true
	hRes := &[]models.HeaderResult{}

	res := &models.Result{
		StatusCode: models.IntResult{
			Normal:   false,
			Expected: tc.HttpResp.StatusCode,
			Actual:   resp.StatusCode,
		},
		BodyResult: []models.BodyResult{{
			Normal:   false,
			Type:     bodyType,
			Expected: tc.HttpResp.Body,
			Actual:   resp.Body,
		}},
	}

	var (
		bodyNoise   []string
		headerNoise = map[string]string{}
	)

	for _, n := range tc.Noise {
		a := strings.Split(n, ".")
		if len(a) > 1 && a[0] == "body" {
			x := strings.Join(a[1:], ".")
			bodyNoise = append(bodyNoise, x)
		} else if a[0] == "header" {
			// if len(a) == 2 {
			//  headerNoise[a[1]] = a[1]
			//  continue
			// }
			headerNoise[a[len(a)-1]] = a[len(a)-1]
			// headerNoise[a[0]] = a[0]
		}
	}

	// stores the json body after removing the noise
	cleanExp, cleanAct := "", ""

	if !pkg.Contains(tc.Noise, "body") && bodyType == models.BodyTypeJSON {
		cleanExp, cleanAct, pass, err = pkg.Match(tc.HttpResp.Body, resp.Body, bodyNoise, r.log)
		if err != nil {
			return false, res, &tc, err
		}
	} else {
		if !pkg.Contains(tc.Noise, "body") && tc.HttpResp.Body != resp.Body {
			pass = false
		}
	}

	res.BodyResult[0].Normal = pass

	if !pkg.CompareHeaders(tc.HttpResp.Header, grpcMock.ToHttpHeader(grpcMock.ToMockHeader(resp.Header)), hRes, headerNoise) {

		pass = false
	}

	res.HeadersResult = *hRes
	if tc.HttpResp.StatusCode == resp.StatusCode {
		res.StatusCode.Normal = true
	} else {

		pass = false
	}
	if !pass {
		logger := pp.New()
		logger.WithLineInfo = false
		logger.SetColorScheme(models.FailingColorScheme)
		var logs = ""

		logs = logs + logger.Sprintf("Testrun failed for testcase with id: %s\n"+
			"Test Result:\n"+
			"\tInput Http Request: %+v\n\n"+
			"\tExpected Response: "+
			"%+v\n\n"+"\tActual Response: "+
			"%+v\n\n"+"DIFF: \n", tc.ID, tc.HttpReq, tc.HttpResp, resp)

		if !res.StatusCode.Normal {
			logs += logger.Sprintf("\tExpected StatusCode: %s"+"\n\tActual StatusCode: %s\n\n", res.StatusCode.Expected, res.StatusCode.Actual)

		}
		var (
			actualHeader   = map[string][]string{}
			expectedHeader = map[string][]string{}
			unmatched      = true
		)

		for _, j := range res.HeadersResult {
			if !j.Normal {
				unmatched = false
				actualHeader[j.Actual.Key] = j.Actual.Value
				expectedHeader[j.Expected.Key] = j.Expected.Value
			}
		}

		if !unmatched {
			logs += "\t Response Headers: {\n"
			for i, j := range expectedHeader {
				logs += logger.Sprintf("\t\t%s"+": {\n\t\t\tExpected value: %+v"+"\n\t\t\tActual value: %+v\n\t\t}\n", i, fmt.Sprintf("%v", j), fmt.Sprintf("%v", actualHeader[i]))
			}
			logs += "\t}\n"
		}
		// TODO: cleanup the logging related code. this is a mess
		if !res.BodyResult[0].Normal {
			logs += "\tResponse body: {\n"
			if json.Valid([]byte(resp.Body)) {
				// compute and log body's json diff
				//diff := cmp.Diff(tc.HttpResp.Body, resp.Body)
				//logs += logger.Sprintf("\t\t%s\n\t\t}\n", diff)
				//expected, actual := pkg.RemoveNoise(tc.HttpResp.Body, resp.Body, bodyNoise, r.log)

				patch, err := jsondiff.Compare(cleanExp, cleanAct)
				if err != nil {
					r.log.Warn("failed to compute json diff", zap.Error(err))
				}
				for _, op := range patch {
					keyStr := op.Path.String()
					if len(keyStr) > 1 && keyStr[0] == '/' {
						keyStr = keyStr[1:]
					}
					logs += logger.Sprintf("\t\t%s"+": {\n\t\t\tExpected value: %+v"+"\n\t\t\tActual value: %+v\n\t\t}\n", keyStr, op.OldValue, op.Value)
				}
				logs += "\t}\n"
			} else {
				// just log both the bodies as plain text without really computing the diff
				logs += logger.Sprintf("{\n\t\t\tExpected value: %+v"+"\n\t\t\tActual value: %+v\n\t\t}\n", tc.HttpResp.Body, resp.Body)

			}
		}
		logs += "--------------------------------------------------------------------\n\n"
		logger.Printf(logs)
	} else {
		logger := pp.New()
		logger.WithLineInfo = false
		logger.SetColorScheme(models.PassingColorScheme)
		var log2 = ""
		log2 += logger.Sprintf("Testrun passed for testcase with id: %s\n\n--------------------------------------------------------------------\n\n", tc.ID)
		logger.Printf(log2)

	}
	return pass, res, &tc, nil
}

func (r *Regression) testGrpc(ctx context.Context, cid, runId, id, app string, resp models.GrpcResp) (bool, *models.Result, *models.TestCase, error) {
	var (
		tc  models.TestCase
		err error
	)
	switch r.testExport {
	case false:
		tc, err = r.tdb.Get(ctx, cid, id)
		if err != nil {
			return false, nil, nil, pkg.LogError("failed to get testcase from DB", r.log, err, map[string]interface{}{"id": id, "cid": cid, "appID": app})
		}
	case true:
		if val, ok := r.yamlTcs.Load(runId); ok {
			tcsMap := val.(sync.Map)
			if val, ok := tcsMap.Load(id); ok {
				tc = val.(models.TestCase)
				// tcsMap.Delete(id)
			} else {
				return false, nil, nil, pkg.LogError("", r.log, fmt.Errorf("failed to load testcase from tcs map coresponding to testcaseId: %s", pkg.SanitiseInput(id)))
			}
		} else {
			return false, nil, nil, pkg.LogError("", r.log, fmt.Errorf("failed to load testcases coresponding to runId: %s", pkg.SanitiseInput(runId)))
		}
	}
	bodyType := models.BodyTypePlain
	if json.Valid([]byte(resp.Body)) {
		bodyType = models.BodyTypeJSON
	}
	pass := true

	res := &models.Result{
		BodyResult: []models.BodyResult{
			{
				Normal:   false,
				Type:     bodyType,
				Expected: tc.GrpcResp.Body,
				Actual:   resp.Body,
			},
			{
				Normal:   true,
				Type:     models.BodyTypeError,
				Expected: tc.GrpcResp.Err,
				Actual:   resp.Err,
			},
		},
	}

	var (
		bodyNoise   []string
		headerNoise = map[string]string{}
	)

	for _, n := range tc.Noise {
		a := strings.Split(n, ".")
		if len(a) > 1 && a[0] == "body" {
			x := strings.Join(a[1:], ".")
			bodyNoise = append(bodyNoise, x)
		} else if a[0] == "header" {
			headerNoise[a[len(a)-1]] = a[len(a)-1]
		}
	}

	// stores the json body after removing the noise
	cleanExp, cleanAct := "", ""

	if !pkg.Contains(tc.Noise, "body") && bodyType == models.BodyTypeJSON {
		cleanExp, cleanAct, pass, err = pkg.Match(tc.GrpcResp.Body, resp.Body, bodyNoise, r.log)
		if err != nil {
			return false, res, &tc, err
		}
	} else {
		if !pkg.Contains(tc.Noise, "body") && tc.GrpcResp.Body != resp.Body {
			pass = false
		}
	}
	res.BodyResult[0].Normal = pass

	if diff := deep.Equal(resp.Err, tc.GrpcResp.Err); diff != nil {
		pass = false
		res.BodyResult[1].Normal = false
	}

	if !pass {
		logger := pp.New()
		logger.WithLineInfo = false
		logger.SetColorScheme(models.FailingColorScheme)
		var logs = ""

		logs = logs + logger.Sprintf("Testrun failed for testcase with id: %s\n"+
			"Test Result:\n"+
			"\tInput Grpc Request: %+v\n\n"+
			"\tExpected Response: "+
			"%+v\n\n"+"\tActual Response: "+
			"%+v\n\n"+"DIFF: \n", tc.ID, tc.GrpcReq, tc.GrpcResp, resp)

		// TODO: cleanup the logging related code. this is a mess
		if !res.BodyResult[0].Normal {
			logs += "\tResponse body: {\n"
			if json.Valid([]byte(resp.Body)) {
				// compute and log body's json diff
				//diff := cmp.Diff(tc.HttpResp.Body, resp.Body)
				//logs += logger.Sprintf("\t\t%s\n\t\t}\n", diff)
				//expected, actual := pkg.RemoveNoise(tc.HttpResp.Body, resp.Body, bodyNoise, r.log)

				patch, err := jsondiff.Compare(cleanExp, cleanAct)
				if err != nil {
					r.log.Warn("failed to compute json diff", zap.Error(err))
				}
				for _, op := range patch {
					keyStr := op.Path.String()
					if len(keyStr) > 1 && keyStr[0] == '/' {
						keyStr = keyStr[1:]
					}
					logs += logger.Sprintf("\t\t%s"+": {\n\t\t\tExpected value: %+v"+"\n\t\t\tActual value: %+v\n\t\t}\n", keyStr, op.OldValue, op.Value)
				}
				logs += "\t}\n"
			} else {
				// just log both the bodies as plain text without really computing the diff
				logs += logger.Sprintf("{\n\t\t\tExpected value: %+v"+"\n\t\t\tActual value: %+v\n\t\t}\n", tc.GrpcResp, resp)

			}
		}
		if !res.BodyResult[1].Normal {
			logs += "\tResponse error: {\n"
			logs += logger.Sprintf("\t\tExpected value: %+v"+"\n\t\tActual value: %+v\n\t}\n", tc.GrpcResp.Err, resp.Err)
		}
		logs += "--------------------------------------------------------------------\n\n"
		logger.Printf(logs)
	} else {
		logger := pp.New()
		logger.WithLineInfo = false
		logger.SetColorScheme(models.PassingColorScheme)
		var log2 = ""
		log2 += logger.Sprintf("Testrun passed for testcase with id: %s\n\n--------------------------------------------------------------------\n\n", tc.ID)
		logger.Printf(log2)

	}
	return pass, res, &tc, nil
}

func (r *Regression) Test(ctx context.Context, cid, app, runID, id, testCasePath, mockPath string, resp models.HttpResp) (bool, error) {
	var t *models.Test
	started := time.Now().UTC()
	ok, res, tc, err := r.test(ctx, cid, runID, id, app, resp)
	if tc != nil {
		t = &models.Test{
			ID:         uuid.New().String(),
			Started:    started.Unix(),
			RunID:      runID,
			TestCaseID: id,
			URI:        tc.URI,
			Req:        tc.HttpReq,
			Dep:        tc.Deps,
			Resp:       resp,
			Result:     *res,
			Noise:      tc.Noise,
		}
	}
	t.Completed = time.Now().UTC().Unix()

	if err != nil {
		pkg.LogError("failed to run the testcase", r.log, err, map[string]interface{}{"app": app, "cid": cid})
		t.Status = models.TestStatusFailed
	}
	t.Status = models.TestStatusFailed
	if ok {
		t.Status = models.TestStatusPassed
	}
	defer func() {

		if r.testExport {
			mockIds := []string{}
			for i := 0; i < len(tc.Mocks); i++ {
				mockIds = append(mockIds, tc.Mocks[i].Name)
			}
			r.testReportFS.Lock()
			r.testReportFS.SetResult(runID, models.TestResult{
				Kind:       models.HTTP,
				Name:       runID,
				Status:     t.Status,
				Started:    t.Started,
				Completed:  t.Completed,
				TestCaseID: id,
				Req: models.MockHttpReq{
					Method:     t.Req.Method,
					ProtoMajor: t.Req.ProtoMajor,
					ProtoMinor: t.Req.ProtoMinor,
					URL:        t.Req.URL,
					URLParams:  t.Req.URLParams,
					Header:     grpcMock.ToMockHeader(t.Req.Header),
					Body:       t.Req.Body,
				},
				Res: models.MockHttpResp{
					StatusCode:    t.Resp.StatusCode,
					Header:        grpcMock.ToMockHeader(t.Resp.Header),
					Body:          t.Resp.Body,
					StatusMessage: t.Resp.StatusMessage,
					ProtoMajor:    t.Resp.ProtoMajor,
					ProtoMinor:    t.Resp.ProtoMinor,
				},
				Mocks:        mockIds,
				TestCasePath: testCasePath,
				MockPath:     mockPath,
				Noise:        tc.Noise,
				Result:       *res,
			})
			r.testReportFS.Lock()
			defer r.testReportFS.Unlock()
		} else {
			err2 := r.saveResult(ctx, t)
			if err2 != nil {
				pkg.LogError("failed test result to db", r.log, err2, map[string]interface{}{"cid": cid, "app": app})
			}
		}
	}()
	return ok, nil
}

func (r *Regression) TestGrpc(ctx context.Context, resp models.GrpcResp, cid, app, runID, id, testCasePath, mockPath string) (bool, error) {
	var t *models.Test
	started := time.Now().UTC()
	ok, res, tc, err := r.testGrpc(ctx, cid, runID, id, app, resp)
	if tc != nil {
		t = &models.Test{
			ID:         uuid.New().String(),
			Started:    started.Unix(),
			RunID:      runID,
			TestCaseID: id,
			// GrpcMethod: tc.GrpcReq.Method,
			GrpcReq:  tc.GrpcReq,
			Dep:      tc.Deps,
			GrpcResp: resp,
			Result:   *res,
			Noise:    tc.Noise,
		}
	}
	t.Completed = time.Now().UTC().Unix()

	if err != nil {
		pkg.LogError("failed to run the grpc testcase", r.log, err, map[string]interface{}{"cid": cid, "app": app})
		t.Status = models.TestStatusFailed
	}
	t.Status = models.TestStatusFailed
	if ok {
		t.Status = models.TestStatusPassed
	}
	defer func() {

		if r.testExport {
			mockIds := []string{}
			for i := 0; i < len(tc.Mocks); i++ {
				mockIds = append(mockIds, tc.Mocks[i].Name)
			}
			r.testReportFS.Lock()
			r.testReportFS.SetResult(runID, models.TestResult{
				Kind:         models.GRPC_EXPORT,
				Name:         runID,
				Status:       t.Status,
				Started:      t.Started,
				Completed:    t.Completed,
				TestCaseID:   id,
				GrpcReq:      tc.GrpcReq,
				GrpcResp:     resp,
				Mocks:        mockIds,
				TestCasePath: testCasePath,
				MockPath:     mockPath,
				Noise:        tc.Noise,
				Result:       *res,
			})
			r.testReportFS.Lock()
			r.testReportFS.Unlock()
		} else {
			err2 := r.saveResult(ctx, t)
			if err2 != nil {
				pkg.LogError("ailed test result to db", r.log, err2, map[string]interface{}{"cid": cid, "app": app})
			}
		}
	}()
	return ok, nil
}

func (r *Regression) saveResult(ctx context.Context, t *models.Test) error {
	err := r.rdb.PutTest(ctx, *t)
	if err != nil {
		return err
	}
	if t.Status == models.TestStatusFailed {
		err = r.rdb.Increment(ctx, false, true, t.RunID)
	} else {
		err = r.rdb.Increment(ctx, true, false, t.RunID)
	}

	if err != nil {
		return err
	}
	return nil
}

func (r *Regression) deNoiseYaml(ctx context.Context, id, path, body string, h http.Header) error {
	tcs, err := r.mockFS.Read(ctx, path, id, false)
	reqType := ctx.Value("reqType").(models.Kind)
	if err != nil {
		return pkg.LogError("failed to read testcase from yaml", r.log, err, map[string]interface{}{"id": id, "path": path})
	}
	if len(tcs) == 0 {
		return pkg.LogError("no testcase exists with", r.log, fmt.Errorf("no testcase to be denoised"), map[string]interface{}{"id": id, "at path": path})
	}
	docs, err := grpcMock.Decode(tcs)
	if err != nil {
		return pkg.LogError("", r.log, err)
	}
	tc := docs[0]
	var oldResp map[string][]string
	switch reqType {
	case models.HTTP:
		oldResp, err = pkg.FlattenHttpResponse(utils.GetStringMap(tc.Spec.Res.Header), tc.Spec.Res.Body)
		if err != nil {
			return pkg.LogError("failed to flatten response", r.log, err)
		}

	case models.GRPC_EXPORT:
		oldResp = map[string][]string{}
		err := pkg.AddHttpBodyToMap(tc.Spec.GrpcResp.Body, oldResp)
		if err != nil {
			return pkg.LogError("failed to flatten response", r.log, err)
		}
	}

	noise := pkg.FindNoisyFields(oldResp, func(k string, v []string) bool {
		var newResp map[string][]string
		switch reqType {
		case models.HTTP:
			newResp, err = pkg.FlattenHttpResponse(h, body)
			if err != nil {
				pkg.LogError("failed to flatten response", r.log, err)
				return false
			}

		case models.GRPC_EXPORT:
			newResp = map[string][]string{}
			err := pkg.AddHttpBodyToMap(body, newResp)
			if err != nil {
				pkg.LogError("failed to flatten response", r.log, err)
				return false
			}
		}
		// TODO : can we simplify this by checking and return false first?
		v2, ok := newResp[k]
		if !ok {
			return true
		}
		if !reflect.DeepEqual(v, v2) {
			return true
		}
		return false

	})
	r.log.Debug("Noise Array : ", zap.Any("", noise))
	tc.Spec.Assertions["noise"] = utils.ToStrArr(noise)
	doc, err := grpcMock.Encode(tc)
	if err != nil {
		return pkg.LogError("", r.log, err)
	}
	enc := doc
	d, err := yaml.Marshal(enc)
	if err != nil {
		return pkg.LogError("failed to marshal document to yaml", r.log, err)
	}
	err = os.WriteFile(filepath.Join(path, id+".yaml"), d, os.ModePerm)
	if err != nil {
		return pkg.LogError("failed to write test to yaml file", r.log, err)
	}

	return nil
}

func (r *Regression) DeNoise(ctx context.Context, cid, id, app, body string, h http.Header, path string) error {

	if r.testExport {
		return r.deNoiseYaml(ctx, id, path, body, h)
	}
	tc, err := r.tdb.Get(ctx, cid, id)
	reqType := ctx.Value("reqType").(models.Kind)
	var tcRespBody string
	switch reqType {
	case models.HTTP:
		tcRespBody = tc.HttpResp.Body

	case models.GRPC_EXPORT:
		tcRespBody = tc.GrpcResp.Body
	}
	if err != nil {
		return pkg.LogError("failed to get testcase from DB", r.log, err, map[string]interface{}{"id": id, "cid": cid, "appID": app})
	}

	a, b := map[string][]string{}, map[string][]string{}

	if reqType == models.HTTP {
		// add headers
		for k, v := range tc.HttpResp.Header {
			a["header."+k] = []string{strings.Join(v, "")}
		}

		for k, v := range h {
			b["header."+k] = []string{strings.Join(v, "")}
		}
	}

	err = pkg.AddHttpBodyToMap(tcRespBody, a)
	if err != nil {
		return pkg.LogError("failed to parse response body", r.log, err, map[string]interface{}{"id": id, "cid": cid, "appID": app})
	}

	err = pkg.AddHttpBodyToMap(body, b)
	if err != nil {
		return pkg.LogError("failed to parse response body", r.log, err, map[string]interface{}{"id": id, "cid": cid, "appID": app})
	}
	// r.log.Debug("denoise between",zap.Any("stored object",a),zap.Any("coming object",b))
	var noise []string
	for k, v := range a {
		v2, ok := b[k]
		if !ok {
			noise = append(noise, k)
			continue
		}
		if !reflect.DeepEqual(v, v2) {
			noise = append(noise, k)
		}
	}
	// r.log.Debug("Noise Array : ",zap.Any("",noise))
	tc.Noise = noise
	err = r.tdb.Upsert(ctx, tc)
	if err != nil {
		return pkg.LogError("failed to update noise fields for testcase", r.log, err, map[string]interface{}{"id": id, "cid": cid, "appID": app})
	}
	return nil
}

func (r *Regression) Normalize(ctx context.Context, cid, id string) error {
	t, err := r.rdb.ReadTest(ctx, id)
	if err != nil {
		pkg.LogError("failed to fetch test from db", r.log, err, map[string]interface{}{"cid": cid, "id": id})
		return errors.New("test not found")
	}
	tc, err := r.tdb.Get(ctx, cid, t.TestCaseID)
	if err != nil {
		pkg.LogError("failed to fetch testcase from db", r.log, err, map[string]interface{}{"cid": cid, "id": id})
		return errors.New("testcase not found")
	}
	// update the responses
	tc.HttpResp = t.Resp
	err = r.tdb.Upsert(ctx, tc)
	if err != nil {
		pkg.LogError("failed to update testcase from db", r.log, err, map[string]interface{}{"cid": cid, "id": id})
		return errors.New("could not update testcase")
	}
	r.tele.Normalize(r.client, ctx)
	return nil
}

func (r *Regression) GetTestRun(ctx context.Context, summary bool, cid string, user, app, id *string, from, to *time.Time, offset *int, limit *int) ([]*models.TestRun, error) {
	off, lim := 0, 25
	if offset != nil {
		off = *offset
	}
	if limit != nil {
		lim = *limit
	}
	res, err := r.rdb.Read(ctx, cid, user, app, id, from, to, off, lim)
	if err != nil {
		pkg.LogError("failed to read test runs from DB", r.log, err, map[string]interface{}{"cid": cid, "user": user, "app": app, "id": id, "from": from, "to": to})
		return nil, errors.New("failed getting test runs")
	}
	err = r.updateStatus(ctx, res)
	if err != nil {
		return nil, err
	}
	if summary {
		return res, nil
	}
	if len(res) == 0 {
		return res, nil
	}

	for _, v := range res {
		tests, err1 := r.rdb.ReadTests(ctx, v.ID)
		if err1 != nil {
			pkg.LogError("failed getting tests from DB", r.log, err1, map[string]interface{}{"test run id": v.ID})
			return nil, errors.New("failed getting tests from DB")
		}
		v.Tests = tests
	}
	return res, nil
}

func (r *Regression) updateStatus(ctx context.Context, trs []*models.TestRun) error {
	tests := 0

	for _, tr := range trs {

		if tr.Status != models.TestRunStatusRunning {
			// r.tele.Testrun(tr.Success, tr.Failure, r.client, ctx)
			tests++
			continue
		}
		tests, err1 := r.rdb.ReadTests(ctx, tr.ID)

		if err1 != nil {
			pkg.LogError("failed getting tests from DB", r.log, err1, map[string]interface{}{"test run id": tr.ID})
			return errors.New("failed getting tests from DB")
		}
		if len(tests) == 0 {

			// check if the testrun is more than 5 mins old
			err := r.failOldTestRuns(ctx, tr.Created, tr)
			if err != nil {
				return err
			}
			continue

		}
		// find the newest test
		ts := tests[0].Started
		for _, test := range tests {
			if test.Started > ts {
				ts = test.Started
			}
		}
		// if the oldest test is older than 5 minutes then fail the whole test run
		err := r.failOldTestRuns(ctx, ts, tr)
		if err != nil {
			return err
		}
	}
	if tests != r.runCount {

		for _, tr := range trs {

			if tr.Status != models.TestRunStatusRunning {

				r.tele.Testrun(tr.Success, tr.Failure, r.client, ctx)
			}
		}
		r.runCount = tests
	}
	return nil
}

func (r *Regression) failOldTestRuns(ctx context.Context, ts int64, tr *models.TestRun) error {
	diff := time.Now().UTC().Sub(time.Unix(ts, 0))
	if diff < 5*time.Minute {
		return nil
	}
	tr.Status = models.TestRunStatusFailed
	err2 := r.rdb.Upsert(ctx, *tr)
	if err2 != nil {
		pkg.LogError("failed validating and updating test run status", r.log, err2, map[string]interface{}{"test run id": tr.ID, "cid": tr.CID})
		return errors.New("failed validating and updating test run status")
	}
	return nil

}

func (r *Regression) PutTest(ctx context.Context, run models.TestRun, testExport bool, runId, testCasePath, mockPath, testReportPath string, totalTcs int) error {
	if run.Status == models.TestRunStatusRunning {
		if testExport {
			err := r.startTestRun(ctx, runId, testCasePath, mockPath, testReportPath, totalTcs)
			if err != nil {
				return err
			}
		}
		pp.SetColorScheme(models.PassingColorScheme)
		pp.Printf("\n <=========================================> \n  TESTRUN STARTED with id: %s\n"+"\tFor App: %s\n"+"\tTotal tests: %s\n <=========================================> \n\n", run.ID, run.App, run.Total)
	} else {
		var (
			total   int
			success int
			failure int
			err     error
		)
		if testExport {
			err = r.stopTestRun(ctx, runId, testReportPath)
			if err != nil {
				return err
			}
			res := models.TestReport{}
			res, err = r.testReportFS.Read(ctx, testReportPath, run.ID)
			total = res.Total
			success = res.Success
			failure = res.Failure
		} else {
			var res *models.TestRun
			res, err = r.rdb.ReadOne(ctx, run.ID)
			total = res.Total
			success = res.Success
			failure = res.Failure
		}
		if err != nil {
			return pkg.LogError("failed to load testrun for logging test summary", r.log, err)
		}
		if run.Status == models.TestRunStatusFailed {
			pp.SetColorScheme(models.FailingColorScheme)
		} else {
			pp.SetColorScheme(models.PassingColorScheme)
		}
		pp.Printf("\n <=========================================> \n  TESTRUN SUMMARY. For testrun with id: %s\n"+"\tTotal tests: %s\n"+"\tTotal test passed: %s\n"+"\tTotal test failed: %s\n <=========================================> \n\n", run.ID, total, success, failure)
	}
	if !testExport {
		return r.rdb.Upsert(ctx, run)
	}
	return nil
}
