package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

func main() {
	// This is needed to make `glog` believe that the flags have already been parsed, otherwise
	// every log messages is prefixed by an error message stating the the flags haven't been
	// parsed.
	_ = flag.CommandLine.Parse([]string{})

	//pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	// Always log to stderr by default
	if err := flag.Set("logtostderr", "true"); err != nil {
		glog.Infof("Unable to set logtostderr to true")
	}

	rootCmd := &cobra.Command{
		Use:  "authz",
		Long: "authz service.",
	}

	//service.Execute()
	if err := rootCmd.Execute(); err != nil {
		glog.Fatalf("error running command: %v", err)
	}

	//TODO Remove later - Helloworld
	http.HandleFunc("/", HelloServer)
	http.HandleFunc("/CheckPermission", CheckPermission)
	http.ListenAndServe(":8080", nil)
}

// TODO - Remove later
func HelloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}

func CheckPermission(w http.ResponseWriter, r *http.Request) {
	var cpr v1.CheckPermissionRequest

	err := json.NewDecoder(r.Body).Decode(&cpr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
