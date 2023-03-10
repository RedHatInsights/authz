package main

import (
	"authz/app"
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
		Use:   "authz",
		Short: "authz service",
		Long:  `authz service`,
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
	//TODO: empty for now.
	var licensedSeats = make(map[string]map[string]bool)
	p1 := impl.StubPrincipalStore{ //TODO: SpiceDbPrincipalStore (and discuss generally)
		Principals: map[string]app.Principal{
			"token": app.NewPrincipal("token", "aspian"),
			"alice": app.NewPrincipal("alice", "aspian"),
			"bob":   app.NewPrincipal("bob", "aspian"),
			"chuck": app.NewPrincipal("chuck", "aspian"),
		},
	}
	var services host.Services
	if !shared.StringEmpty(endpoint) && !shared.StringEmpty(token) {
		authzclient := authzed.NewAuthzedConnection(endpoint, token)
		authz := impl.SpiceDBAuthzStore{
			Authzed:       authzclient,
			LicensedSeats: licensedSeats,
		}
		if shared.StringEqualsIgnoreCase(store, "spicedb") {
			services = host.Services{
				Authz:      authz,
				Licensing:  authz,
				Principals: &p1,
			}
			// Added below line to test the implementation - commented out for now, since testing is done
			//authzclient.ReadSchema()
		}

	} else {
		authz := impl.StubAuthzStore{
			AuthzdUsers:   map[string]bool{"token": true, "alice": true, "bob": true, "chuck": false},
			LicensedSeats: make(map[string]map[string]bool),
		}

		p2 := impl.StubPrincipalStore{
			Principals: map[string]app.Principal{
				"token": app.NewPrincipal("token", "aspian"),
				"alice": app.NewPrincipal("alice", "aspian"),
				"bob":   app.NewPrincipal("bob", "aspian"),
				"chuck": app.NewPrincipal("chuck", "aspian"),
			},
		}

		services = host.Services{
			Authz:      &authz,
			Licensing:  &authz,
			Principals: &p2,
		}
	}

	wait := sync.WaitGroup{}
	web := host.NewWeb(services)
	gRPC := host.NewGrpcServer(services)

	wait.Add(2)
	go web.Host(&wait, gRPC, gRPC)
	go gRPC.Host(&wait)

	wait.Wait()
}
