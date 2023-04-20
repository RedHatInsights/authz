package http

import "net/http"

type middleware func(http.Handler) http.Handler
type chain []middleware

func createChain(middlewares ...middleware) chain {
	var slice chain
	return append(slice, middlewares...)
}

func (c chain) then(originalHandler http.Handler) http.Handler {
	for i := range c {
		// Same as to middleware1(middleware2(m3(originalHandler)))
		originalHandler = c[len(c)-1-i](originalHandler)
	}
	return originalHandler
}
