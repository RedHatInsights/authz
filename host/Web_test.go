package host

import (
	"authz/app"
	"authz/host/impl"
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
	resp := runRequest(post("/v1alpha/permissions/check", "bad",
		`{"subject": "okay", "operation": "do_stuff", "resourcetype": "Thing", "resourceid": "1"}`))

	assert.Equal(t, 403, resp.StatusCode)
}

func TestCheckErrorsWhenTokenMissing(t *testing.T) {
	t.Parallel()
	resp := runRequest(post("/v1alpha/permissions/check", "",
		`{"subject": "okay", "operation": "do_stuff", "resourcetype": "Thing", "resourceid": "1"}`))

	assert.Equal(t, 401, resp.StatusCode)
}

func TestCheckReturnsTrueWhenUserAuthorized(t *testing.T) {
	t.Parallel()
	resp := runRequest(post("/v1alpha/permissions/check", "system",
		`{"subject": "okay", "operation": "do_stuff", "resourcetype": "Thing", "resourceid": "1"}`))

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, true)
}

func TestCheckReturnsFalseWhenUserNotAuthorized(t *testing.T) {
	t.Parallel()
	resp := runRequest(post("/v1alpha/permissions/check", "system",
		`{"subject": "bad", "operation": "do_stuff", "resourcetype": "Thing", "resourceid": "1"}`))

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, false)
}

func TestAssignLicenseReturnsSuccess(t *testing.T) {
	t.Parallel()
	resp := runRequest(post("/v1alpha/license/seats", "okay",
		`{
			"tenantId": "aspian",
			"subjects": [
			  "okay"
			],
			"serviceId": "wisdom"
		  }`))

	assertJSONResponse(t, resp, 200, `{}`)
}

func TestUnassignLicenseReturnsSuccess(t *testing.T) {
	t.Parallel()
	resp := runRequest(delete("/v1alpha/license/seats", "okay",
		`{
			"tenantId": "aspian",
			"subjects": [
			  "okay"
			],
			"serviceId": "wisdom"
		  }`))

	assertJSONResponse(t, resp, 200, `{}`)
}

func TestGrantedLicenseAllowsUse(t *testing.T) {
	t.Parallel()
	srv := createTestServer()

	//The user isn't licensed initially, use is denied
	resp := runRequestWithServer(post("/v1alpha/permissions/check", "system",
		`{"subject": "okay", "operation": "use", "resourcetype": "service", "resourceid": "wisdom"}`), srv)

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, false)

	//Grant a license
	resp = runRequestWithServer(post("/v1alpha/license/seats", "okay",
		`{
		"tenantId": "aspian",
		"subjects": [
		  "okay"
		],
		"serviceId": "wisdom"
	  }`), srv)

	assertJSONResponse(t, resp, 200, `{}`)

	//Should be allowed now
	resp = runRequestWithServer(post("/v1alpha/permissions/check", "system",
		`{"subject": "okay", "operation": "use", "resourcetype": "service", "resourceid": "wisdom"}`), srv)

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, true)
}

func post(uri string, token string, body string) *http.Request {
	return reqWithBody(http.MethodPost, uri, token, body)
}

func delete(uri string, token string, body string) *http.Request {
	return reqWithBody(http.MethodDelete, uri, token, body)
}

func reqWithBody(method string, uri string, token string, body string) *http.Request {
	req := httptest.NewRequest(method, uri, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	if token != "" {
		req.Header.Set("Authorization", token)
	}
	return req
}

// runRequest connects a mock HTTP front-end to a mock Store back-end using the generated proxy code and real gRPC implementation to test that integration
func runRequest(req *http.Request) *http.Response {
	srv := createTestServer()

	return runRequestWithServer(req, srv)
}

func createTestServer() *GrpcServer {
	authz := mockAuthzStore()
	principals := mockPrincipalStore()

	return NewGrpcServer(Services{Authz: authz, Licensing: authz, Principals: principals})
}

func runRequestWithServer(req *http.Request, srv *GrpcServer) *http.Response {
	mux, _ := createMultiplexer(srv, srv)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec.Result()
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

func mockAuthzStore() impl.StubAuthzStore {
	return impl.StubAuthzStore{
		AuthzdUsers: map[string]bool{
			"system": true,
			"okay":   true,
			"bad":    false,
		},
		LicensedSeats: make(map[string]map[string]bool),
	}
}

func mockPrincipalStore() impl.StubPrincipalStore {
	return impl.StubPrincipalStore{
		Principals: map[string]app.Principal{
			"system": app.NewPrincipal("system", "wisdom"),
			"okay":   app.NewPrincipal("okay", "aspian"),
			"bad":    app.NewPrincipal("bad", "aspian"),
		},
	}
}
