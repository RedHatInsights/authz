package main

import (
	authzed "authz/app/client/authzed"
	"authz/app/shared"
	"authz/flags"
	"authz/host"
	"authz/host/impl"
	"flag"
	"sync"

	"github.com/golang/glog"
	cobra "github.com/spf13/cobra"
)

func main() {

	//TODO: factor all this out behind a logging abstraction
	// This is needed to make `glog` believe that the flags have already been parsed, otherwise
	// every log messages is prefixed by an error message stating the flags haven't been
	// parsed.
	_ = flag.CommandLine.Parse([]string{})

	//pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	// Always log to stderr by default
	if err := flag.Set("logtostderr", "true"); err != nil {
		glog.Infof("Unable to set logtostderr to true")
	}

	var rootCmd = &cobra.Command{
		Use:   "ciam-authzed-api",
		Short: "ciam-authzed-api",
		Long:  `ciam-authzed-api serve`,
		Run:   Serve,
	}
	rootCmd.Flags().String("endpoint", "", "endpoint")
	rootCmd.Flags().String("token", "", "token")
	rootCmd.Flags().String("store", "stub", "stub or spicedb")

	if err := rootCmd.Execute(); err != nil {
		glog.Fatalf("error running command: %v", err)
	}

}

// Serve - Serve command
func Serve(cmd *cobra.Command, args []string) {
	endpoint := flags.MustGetString("endpoint", cmd.Flags())
	token := flags.MustGetString("token", cmd.Flags())
	store := flags.MustGetString("store", cmd.Flags())

	var services host.Services
	if !shared.StringEmpty(endpoint) && !shared.StringEmpty(token) {
		authzclient := authzed.NewAuthzedConnection(endpoint, token)
		if shared.StringEqualsIgnoreCase(store, "spicedb") {
			services = host.Services{Store: impl.AuthzStore{Authzed: authzclient}}
			// Added below line to test the implementation - commented out for now, since testing is done
			//authzclient.ReadSchema()
		}

	} else {
		services = host.Services{Store: impl.StubAuthzStore{Data: map[string]bool{
			"token": true,
			"alice": true,
			"bob":   true,
			"chuck": false,
		}}}
	}

	wait := sync.WaitGroup{}
	web := host.NewWeb(services)
	gRPC := host.NewGrpcServer(services)

	wait.Add(2)
	go web.Host(&wait, gRPC)
	go gRPC.Host(&wait)

	wait.Wait()
}
