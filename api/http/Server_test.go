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
	t.SkipNow()
	t.Parallel()
	resp := runRequest(post("/v1alpha/check", "bad",
		`{"subject": "good", "operation": "op", "resourcetype": "Feature", "resourceid": "smarts"}`))

	assert.Equal(t, 403, resp.StatusCode)
}

func TestCheckErrorsWhenTokenMissing(t *testing.T) {
	t.Parallel()
	resp := runRequest(post("/v1alpha/check", "",
		`{"subject": "good", "operation": "op", "resourcetype": "Feature", "resourceid": "smarts"}`))

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

func TestCors_NotImplementedMethod(t *testing.T) {
	t.Parallel()
	srv := createTestServer()

	body := `{
			"assign": [
			  "okay"
			]
		  }`

	req := httptest.NewRequest(http.MethodTrace, "/v1alpha/orgs/aspian/licenses/smarts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "okay")

	resp := runRequestWithServer(req, srv)
	assert.Equal(t, resp.StatusCode, 501)
}

func TestCors_AllowAllOrigins(t *testing.T) {
	t.Parallel()
	srv := createTestServer()

	body := `{
			"assign": [
			  "okay"
			]
		  }`

	req := httptest.NewRequest(http.MethodPost, "/v1alpha/orgs/aspian/licenses/smarts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "okay")
	req.Header.Set("AllowAllOrigins", "true")

	resp := runRequestWithServer(req, srv)
	assert.Equal(t, resp.Header.Get("Vary"), "Origin")
	assertJSONResponse(t, resp, 200, `{}`)
}

func TestGrantedLicenseAffectsCountsAndDetails(t *testing.T) {
	t.Parallel()
	srv := createTestServer()

	//No one is licensed initially, expect a fixed count and none in use
	resp := runRequestWithServer(get("/v1alpha/orgs/aspian/licenses/smarts", "token"), srv)
	assertJSONResponse(t, resp, 200, `{"seatsAvailable":2, "seatsTotal": 2}`)
	resp = runRequestWithServer(get("/v1alpha/orgs/aspian/licenses/smarts/seats", "token"), srv)
	assertJSONResponse(t, resp, 200, `{"users": []}`)

	//Grant a license
	_ = runRequestWithServer(post("/v1alpha/orgs/aspian/licenses/smarts", "okay",
		`{
			"assign": [
			  "okay"
			]
		  }`), srv)

	resp = runRequestWithServer(get("/v1alpha/orgs/aspian/licenses/smarts", "token"), srv)
	assertJSONResponse(t, resp, 200, `{"seatsAvailable":1, "seatsTotal": 2}`)
	resp = runRequestWithServer(get("/v1alpha/orgs/aspian/licenses/smarts/seats", "token"), srv)
	assertJSONResponse(t, resp, 200, `{"users": [{"assigned":true,"displayName":"Okay User","id":"okay"}]}`)
	resp = runRequestWithServer(get("/v1alpha/orgs/aspian/licenses/smarts/seats?filter=assignable", "token"), srv)
	assertJSONResponse(t, resp, 200, `{"users":[{"assigned":false,"displayName":"Bad User","id":"bad"}]}`)
}

func TestOverAssigningLicensesFails(t *testing.T) {
	t.Parallel()

	resp := runRequest(post("/v1alpha/orgs/aspian/licenses/smarts", "okay",
		`{
		"assign": [
			"user1",
			"user2",
			"user3"
		]
	}`))

	assertJSONResponse(t, resp, 400, `{"code": 9, "message": "<<PRESENCE>>", "details": [{"@type":"<<PRESENCE>>", "seatsTotal":%d, "seatsAvailable":%d}]}`, 2, 2) //Two still available because none were consumed
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
	return &mock.StubAccessRepository{Data: map[domain.SubjectID]bool{
		"system": true,
		"okay":   true,
		"bad":    false,
	},
		LicensedSeats: map[string]map[domain.SubjectID]bool{},
		Licenses: map[string]domain.License{
			"smarts": *domain.NewLicense("aspian", "smarts", 2, 0),
		},
	}
}

func mockPrincipalRepository() contracts.PrincipalRepository {
	return &mock.StubPrincipalRepository{
		Principals: map[domain.SubjectID]domain.Principal{
			"system": domain.NewPrincipal("system", "System User", "smarts"),
			"okay":   domain.NewPrincipal("okay", "Okay User", "aspian"),
			"bad":    domain.NewPrincipal("bad", "Bad User", "aspian"),
		},
		DefaultOrg: "aspian",
	}
}
