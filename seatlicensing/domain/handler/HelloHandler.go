package handler

import (
	"fmt"
	"net/http"
)

// GetHello returns a "Hello" with word from path
func GetHello(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
	if err != nil {
		return
	}
}
