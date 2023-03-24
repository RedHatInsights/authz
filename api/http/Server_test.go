package http

import (
	"authz/api/grpc"
	"authz/application"
	"authz/domain/contracts"
	"authz/domain/model"
	"authz/infrastructure/repository/mock"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kinbiko/jsonassert"
	"github.com/stretchr/testify/assert"
)

func TestCheckErrorsWhenCallerNotAuthorized(t *testing.T) {
	t.Parallel()
	resp := runRequest(post("/v1alpha/check", "bad",
		`{"subject": "okay", "operation": "op", "resourcetype": "Feature", "resourceid": "smarts"}`))

	assert.Equal(t, 403, resp.StatusCode)
}

func TestCheckErrorsWhenTokenMissing(t *testing.T) {
	t.Parallel()
	resp := runRequest(post("/v1alpha/check", "",
		`{"subject": "okay", "operation": "op", "resourcetype": "Feature", "resourceid": "smarts"}`))

	assert.Equal(t, 401, resp.StatusCode)
}

func TestCheckReturnsTrueWhenUserAuthorized(t *testing.T) {
	t.Parallel()
	resp := runRequest(post("/v1alpha/check", "system",
		`{"subject": "okay", "operation": "op", "resourcetype": "Feature", "resourceid": "smarts"}`))

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, true)
}

func TestCheckReturnsFalseWhenUserNotAuthorized(t *testing.T) {
	t.Parallel()
	resp := runRequest(post("/v1alpha/check", "system",
		`{"subject": "bad", "operation": "op", "resourcetype": "Feature", "resourceid": "smarts"}`))

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, false)
}

func TestAssignLicenseReturnsSuccess(t *testing.T) {
	t.Parallel()
	resp := runRequest(post("/v1alpha/orgs/aspian/licenses/smarts", "okay",
		`{
			"assign": [
			  "okay"
			]
		  }`))

	assertJSONResponse(t, resp, 200, `{}`)
}

func TestUnassignLicenseReturnsSuccess(t *testing.T) {
	t.Parallel()
	resp := runRequest(post("/v1alpha/orgs/aspian/licenses/smarts", "okay",
		`{
			"unassign": [
			  "okay"
			]
		}`))

	assertJSONResponse(t, resp, 200, `{}`)
}

func TestGrantedLicenseAllowsUse(t *testing.T) {
	t.Parallel()
	srv := createTestServer()

	//The user isn't licensed initially, use is denied
	resp := runRequestWithServer(post("/v1alpha/check", "system",
		`{"subject": "okay", "operation": "use", "resourcetype": "service", "resourceid": "smarts"}`), srv)

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, false)

	//Grant a license
	resp = runRequestWithServer(post("/v1alpha/orgs/aspian/licenses/smarts", "okay",
		`{
			"assign": [
			  "okay"
			]
		  }`), srv)

	assertJSONResponse(t, resp, 200, `{}`)

	//Should be allowed now
	resp = runRequestWithServer(post("/v1alpha/check", "system",
		`{"subject": "okay", "operation": "use", "resourcetype": "service", "resourceid": "smarts"}`), srv)

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, true)
}

func post(uri string, token string, body string) *http.Request {
	return reqWithBody(http.MethodPost, uri, token, body)
}

func runRequest(req *http.Request) *http.Response {
	srv := createTestServer()

	return runRequestWithServer(req, srv)
}

func createTestServer() *grpc.Server {
	accessRepo := mockAccessRepository()
	licenseRepo, _ := accessRepo.(contracts.SeatLicenseRepository)
	principalRepo := mockPrincipalRepository()

	return &grpc.Server{
		AccessAppService:  application.NewAccessAppService(&accessRepo, principalRepo),
		LicenseAppService: application.NewLicenseAppService(accessRepo, licenseRepo, principalRepo),
	}
}

func runRequestWithServer(req *http.Request, srv *grpc.Server) *http.Response {
	mux, _ := createMultiplexer(srv, srv)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec.Result()
}

func reqWithBody(method string, uri string, token string, body string) *http.Request {
	req := httptest.NewRequest(method, uri, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	if token != "" {
		req.Header.Set("Authorization", token)
	}
	return req
}

func assertJSONResponse(t *testing.T, resp *http.Response, statusCode int, template string, args ...interface{}) {
	if assert.NotNil(t, resp) {
		assert.Equal(t, statusCode, resp.StatusCode)

		payload := new(strings.Builder)
		_, err := io.Copy(payload, resp.Body)
		assert.NoError(t, err)

		ja := jsonassert.New(t)
		ja.Assertf(payload.String(), template, args...)
	}
}

func mockAccessRepository() contracts.AccessRepository {
	mock := &mock.StubAccessRepository{Data: map[string]bool{
		"system": true,
		"okay":   true,
		"bad":    false,
	}, LicensedSeats: map[string]map[string]model.License{}}

	mock.UpdateLicense(model.NewLicense("aspian", "smarts", 1, []string{}))

	return mock
}

func mockPrincipalRepository() contracts.PrincipalRepository {
	return &mock.StubPrincipalRepository{
		Principals: map[string]model.Principal{
			"system": model.NewPrincipal("system", "smarts"),
			"okay":   model.NewPrincipal("okay", "aspian"),
			"bad":    model.NewPrincipal("bad", "aspian"),
		},
	}
}
