// Package main starts the app from the app package.
package main

import (
	"authz/app"
	"flag"
	"fmt"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

	var rootCmd = &cobra.Command{
		Use:   "authz",
		Short: "authz service, alpha.",
		Long:  `authz service based on zanzibar access systems. alpha`,
		Run:   serve,
	}

	rootCmd.PersistentFlags().StringP("config", "c", "", "config")

	if err := rootCmd.Execute(); err != nil {
		glog.Fatalf("error running command: %v", err)
	}

}

func serve(cmd *cobra.Command, _ []string) {
	configPath := nonEmptyStringFlag("config", cmd.Flags())
	glog.Infof("Starting authz service with config from: %v", configPath)

	app.Run(configPath)
}

// mustGetDefinedString attempts to get a non-empty string flag from the provided flag set or panic
func nonEmptyStringFlag(flagName string, flags *pflag.FlagSet) string {
	flagVal := mustGetString(flagName, flags)
	if flagVal == "" {
		glog.Fatal(undefinedValueMessage(flagName))
	}
	return flagVal
}

// mustGetString attempts to get a string flag from the provided flag set or panic
func mustGetString(flagName string, flags *pflag.FlagSet) string {
	flagVal, err := flags.GetString(flagName)
	if err != nil {
		glog.Fatalf(notFoundMessage(flagName, err))
	}
	return flagVal
}

func undefinedValueMessage(flagName string) string {
	return fmt.Sprintf("flag %s has undefined value", flagName)
}

func notFoundMessage(flagName string, err error) string {
	return fmt.Sprintf("could not get flag %s from flag set: %s", flagName, err.Error())
}
