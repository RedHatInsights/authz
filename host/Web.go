package host

import (
	"authz/app"
	"authz/app/contracts"
	"authz/app/controllers"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
)

type Web struct {
	services Services
}

func (web Web) Host(wait *sync.WaitGroup) {
	http.HandleFunc("/CheckPermission", web.checkPermission)
	http.ListenAndServe(":8080", nil)

	wait.Done()
}

func (web Web) checkPermission(w http.ResponseWriter, r *http.Request) {
	var webReq CheckWebRequest

	err := json.NewDecoder(r.Body).Decode(&webReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	requestor := r.Header["Authorization"][0]

	req := contracts.CheckRequest{
		Request: contracts.Request{
			Requestor: app.Principal{Id: requestor},
		},
		Subject:   app.Principal{Id: webReq.Subject},
		Operation: webReq.Operation,
		Resource:  app.Resource{Type: webReq.ResourceType, Id: webReq.ResourceId},
	} //TODO: clean up mapping from web contract to inner models. Meat of the method follows.

	action := controllers.NewAccess(web.services.Store)

	result, err := action.Check(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(200)
	w.Write([]byte(strconv.FormatBool(result)))
}

func NewWeb(services Services) Web {
	return Web{services: services}
}

type CheckWebRequest struct {
	Subject      string
	Operation    string
	ResourceType string
	ResourceId   string
}
