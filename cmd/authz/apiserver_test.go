// run test: go test -v
package helloworldService

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHelloworldService(t *testing.T) {
	r := httptest.NewRequest("GET", "/AuthZ", nil)
	w := httptest.NewRecorder()
	h := http.HandlerFunc(HelloServer)

	h.ServeHTTP(w, r)

	// check HTTP response status code
	if s := w.Code; s != http.StatusOK {
		t.Errorf(" Response code error: got %v, want %v", s, http.StatusOK)
	}

	// check HTTP response body
	want := "Hello, AuthZ!"
	if w.Body.String() != want {
		t.Errorf("Response body error: got %v, want %v", w.Body.String(), want)
	}
}
