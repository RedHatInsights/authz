package http

import (
	"authz/api/grpc"
	"authz/application"
	"authz/domain"
	"authz/domain/contracts"
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
	t.SkipNow() //Skip until meta-authz is in place
	t.Parallel()
	resp := runRequest(post("/v1alpha/check", "bad",
		`{"subject": "u1", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}`))

	assert.Equal(t, 403, resp.StatusCode)
}

func TestCheckErrorsWhenTokenMissing(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	resp := runRequest(post("/v1alpha/check", "",
		`{"subject": "u1", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}`))

	assert.Equal(t, 401, resp.StatusCode)
}

func TestCheckReturnsTrueWhenUserAuthorized(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	resp := runRequest(post("/v1alpha/check", "system",
		`{"subject": "u1", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}`))

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, true)
}

func TestCheckReturnsFalseWhenUserNotAuthorized(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	resp := runRequest(post("/v1alpha/check", "system",
		`{"subject": "not_authorized", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}`))

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, false)
}

func TestAssignLicenseReturnsSuccess(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	resp := runRequest(post("/v1alpha/orgs/o1/licenses/smarts", "system",
		`{
			"assign": [
			  "okay"
			]
		  }`))

	assertJSONResponse(t, resp, 200, `{}`)
}

func TestUnassignLicenseReturnsSuccess(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	resp := runRequest(post("/v1alpha/orgs/o1/licenses/smarts", "system",
		`{
			"unassign": [
			  "u1"
			]
		}`))

	assertJSONResponse(t, resp, 200, `{}`)
}

func TestGrantedLicenseAllowsUse(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	srv := createTestServer()

	//The user isn't licensed initially, use is denied
	resp := runRequestWithServer(post("/v1alpha/check", "system",
		`{"subject": "u2", "operation": "assigned", "resourcetype": "license_seats", "resourceid": "o1/smarts"}`), srv)

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, false)

	//Grant a license
	resp = runRequestWithServer(post("/v1alpha/orgs/o1/licenses/smarts", "okay",
		`{
		"assign": [
			"u2"
			]
			}`), srv)

	assertJSONResponse(t, resp, 200, `{}`)

	spicedbContainer.WaitForQuantizationInterval()

	//Should be allowed now
	resp = runRequestWithServer(post("/v1alpha/check", "system",
		`{"subject": "u2", "operation": "assigned", "resourcetype": "license_seats", "resourceid": "o1/smarts"}`), srv)

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, true)
}

func TestCors_NotImplementedMethod(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	srv := createTestServer()

	body := `{
			"assign": [
			  "okay"
			]
		  }`

	req := httptest.NewRequest(http.MethodTrace, "/v1alpha/orgs/o1/licenses/smarts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "okay")

	resp := runRequestWithServer(req, srv)
	assert.Equal(t, resp.StatusCode, 501)
}

func TestCors_AllowAllOrigins(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	srv := createTestServer()

	body := `{
			"assign": [
			  "okay"
			]
		  }`

	req := httptest.NewRequest(http.MethodPost, "/v1alpha/orgs/o1/licenses/smarts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "okay")
	req.Header.Set("AllowAllOrigins", "true")

	resp := runRequestWithServer(req, srv)
	assert.Equal(t, resp.Header.Get("Vary"), "Origin")
	assertJSONResponse(t, resp, 200, `{}`)
}

func TestGrantedLicenseAffectsCountsAndDetails(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	srv := createTestServer()

	//No one is licensed initially, expect a fixed count and none in use
	resp := runRequestWithServer(get("/v1alpha/orgs/o1/licenses/smarts", "system"), srv)
	assertJSONResponse(t, resp, 200, `{"seatsAvailable":9, "seatsTotal": 10}`)
	resp = runRequestWithServer(get("/v1alpha/orgs/o1/licenses/smarts/seats", "system"), srv)
	assertJSONResponse(t, resp, 200, `{"users": [{"assigned":true,"displayName":"User u1","id":"u1"}]}`)

	//Grant a license
	_ = runRequestWithServer(post("/v1alpha/orgs/o1/licenses/smarts", "okay",
		`{
			"assign": [
			  "okay"
			]
		  }`), srv)
	spicedbContainer.WaitForQuantizationInterval()

	resp = runRequestWithServer(get("/v1alpha/orgs/o1/licenses/smarts", "system"), srv)
	assertJSONResponse(t, resp, 200, `{"seatsAvailable":8, "seatsTotal": 10}`)
	resp = runRequestWithServer(get("/v1alpha/orgs/o1/licenses/smarts/seats", "system"), srv)
	assertJSONResponse(t, resp, 200, `{"users": ["<<UNORDERED>>", {"assigned":true,"displayName":"Okay User","id":"okay"}, {"assigned":true,"displayName":"User u1","id":"u1"}]}`)
	resp = runRequestWithServer(get("/v1alpha/orgs/o1/licenses/smarts/seats?filter=assignable", "token"), srv)
	assertJSONResponse(t, resp, 200, `{"users":["<<UNORDERED>>", {"assigned":false,"displayName":"System User","id":"system"},{"assigned":false,"displayName":"Bad User","id":"bad"}]}`)
}

func TestOverAssigningLicensesFails(t *testing.T) {
	t.SkipNow()
	t.Parallel()

	resp := runRequest(post("/v1alpha/orgs/o1/licenses/smarts", "okay",
		`{
		"assign": [
			"user1",
			"user2",
			"user3",
			"user4",
			"user5",
			"user6",
			"user7",
			"user8",
			"user9",
			"user10"
		]
	}`))

	assert.Equal(t, 400, resp.StatusCode)
}

func post(uri string, token string, body string) *http.Request {
	return reqWithBody(http.MethodPost, uri, token, body)
}

func get(uri string, token string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, uri, strings.NewReader(""))
	if token != "" {
		req.Header.Add("Authorization", token)
	}
	return req
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
		LicenseAppService: application.NewLicenseAppService(&accessRepo, &licenseRepo, principalRepo),
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
	client, err := spicedbContainer.CreateClient()
	if err != nil {
		panic(err)
	}

	return client
}

func mockPrincipalRepository() contracts.PrincipalRepository {
	return &mock.StubPrincipalRepository{
		Principals: map[domain.SubjectID]domain.Principal{
			"system": domain.NewPrincipal("system", "System User", "o1"),
			"okay":   domain.NewPrincipal("okay", "Okay User", "o1"),
			"bad":    domain.NewPrincipal("bad", "Bad User", "o1"),
		},
		DefaultOrg: "o1",
	}
}
