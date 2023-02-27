package host

import (
	"authz/app"
	"authz/app/contracts"
	"authz/app/controllers"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/golang/glog"
)

type Web struct {
	services Services
}

func (web Web) Host(wait *sync.WaitGroup) {
	http.HandleFunc("/CheckPermission", web.checkPermission)

	if _, err := os.Stat("/etc/tls/tls.crt"); err == nil {
		if _, err := os.Stat("/etc/tls/tls.key"); err == nil { //Cert and key exisits start server in HTTPS mode
			glog.Info("TLS cert and Key found  - Starting server in secure HTTPs mode")

			_ = http.ListenAndServeTLS(":8443", "/etc/tls/tls.crt", "/etc/tls/tls.key", nil)
		}
	} else { // For all cases of error - we start a plain HTTP server
		glog.Info("TLS cert or Key not found  - Starting server in unsercure plain HTTP mode")
		_ = http.ListenAndServe(":8080", nil)
	}

	wait.Done()
}

func (web Web) checkPermission(w http.ResponseWriter, r *http.Request) {
	var webReq CheckWebRequest

	err := json.NewDecoder(r.Body).Decode(&webReq)
	if err != nil {
		glog.Errorf("Error decoding payload: %s", err)
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
		glog.Errorf("Error processing request: %s", err)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(200)
	_, err = w.Write([]byte(strconv.FormatBool(result)))

	if err != nil {
		glog.Errorf("Error sending response: %s", err)
	}
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
