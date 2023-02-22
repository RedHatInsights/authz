// run test: go test -v
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
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
	want := "Hello, AuthZ! CI/CD deployed this change!"
	if w.Body.String() != want {
		t.Errorf("Response body error: got %v, want %v", w.Body.String(), want)
	}
}

func TestCheckPermission(t *testing.T) {
	cpr := &v1.CheckPermissionRequest{}
	reqCpr, err := json.Marshal(cpr)
	if err != nil {
		t.Errorf("CheckPermissionRequest could not be marshalled.")
	}

	r := httptest.NewRequest("POST", "/CheckPermission", bytes.NewReader(reqCpr))
	w := httptest.NewRecorder()
	h := http.HandlerFunc(CheckPermission)

	h.ServeHTTP(w, r)

	// check HTTP response status code
	if s := w.Code; s != http.StatusOK {
		t.Errorf(" Response code error: got %v, want %v", s, http.StatusOK)
	}
}
