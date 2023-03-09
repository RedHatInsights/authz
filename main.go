package main

import (
	"authz/app"
	"authz/host"
	"authz/host/impl"
	"flag"
	"sync"

	"github.com/golang/glog"
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

	authz := impl.StubAuthzStore{
		AuthzdUsers: map[string]bool{"token": true, "alice": true, "bob": true, "chuck": false},
	}
	principals := impl.StubPrincipalStore{
		Principals: map[string]app.Principal{
			"token": app.NewPrincipal("token", "aspian"),
			"alice": app.NewPrincipal("alice", "aspian"),
			"bob":   app.NewPrincipal("bob", "aspian"),
			"chuck": app.NewPrincipal("chuck", "aspian"),
		},
	}

	services := host.Services{
		Authz:      &authz,
		Licensing:  &authz,
		Principals: &principals,
	}

	wait := sync.WaitGroup{}
	web := host.NewWeb(services)
	gRPC := host.NewGrpcServer(services)

	wait.Add(2)
	go web.Host(&wait, gRPC, gRPC)
	go gRPC.Host(&wait)

	wait.Wait()
}
