package host

import (
	"authz/app/dependencies"
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
	resp := runRequest(post("/v1alpha/check", "other system",
		`{"subject": "good", "operation": "use", "resourcetype": "Feature", "resourceid": "Wisdom"}`))

	assert.Equal(t, 403, resp.StatusCode)
}

func TestCheckErrorsWhenTokenMissing(t *testing.T) {
	t.Parallel()
	resp := runRequest(post("/v1alpha/check", "",
		`{"subject": "good", "operation": "use", "resourcetype": "Feature", "resourceid": "Wisdom"}`))

	assert.Equal(t, 401, resp.StatusCode)
}

func TestCheckReturnsTrueWhenUserAuthorized(t *testing.T) {
	t.Parallel()
	resp := runRequest(post("/v1alpha/check", "system",
		`{"subject": "okay", "operation": "use", "resourcetype": "Feature", "resourceid": "Wisdom"}`))

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, true)
}

func TestCheckReturnsFalseWhenUserNotAuthorized(t *testing.T) {
	t.Parallel()
	resp := runRequest(post("/v1alpha/check", "system",
		`{"subject": "bad", "operation": "use", "resourcetype": "Feature", "resourceid": "Wisdom"}`))

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, false)
}

func post(uri string, token string, body string) *http.Request {
	return reqWithBody(http.MethodPost, uri, token, body)
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
	mux, _ := createMultiplexer(NewGrpcServer(Services{Store: mockAuthzStore()}))
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

func mockAuthzStore() dependencies.AuthzStore {
	return impl.StubAuthzStore{Data: map[string]bool{
		"system": true,
		"okay":   true,
		"bad":    false,
	}}
}
