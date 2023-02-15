package main

import (
	"flag"
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
}
