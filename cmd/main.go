// Package main starts the app from the app package.
package main

import (
	"authz/app"
	"flag"
	"github.com/golang/glog"
)

// main bootstrapping the current composition of the service
func main() {

	// Needed to make `glog` believe that the flags have already been parsed, otherwise
	// every log message is prefixed by an error message stating the flags haven't been
	// parsed.
	_ = flag.CommandLine.Parse([]string{})

	// Always log to stderr by default
	if err := flag.Set("logtostderr", "true"); err != nil {
		glog.Warningf("Unable to log to stderr by default. using stdout.")
	}

	app.Run()
}
