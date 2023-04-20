package http

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var request = &http.Request{Host: ""}

var middleware1 = func(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Host += "1st_thing"
		h.ServeHTTP(w, request)
	})
}

var middleware2 = func(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Host += "2nd_thing"
		h.ServeHTTP(w, request)
	})
}

var mainHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	r.Host += "Last_thing"
})

func TestCreateChainWithThen(t *testing.T) {
	c := createChain(middleware1, middleware2)
	h := c.then(mainHandler)

	h.ServeHTTP(nil, request)

	assert.Equal(t, "1st_thing2nd_thingLast_thing", request.Host)
}
