// Package main starts the bootstrap from the bootstrap package.
package main

import (
	"authz/bootstrap"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// main bootstrapping the current composition of the service
func main() {

	// Always log to stderr by default
	if err := flag.Set("logtostderr", "true"); err != nil {
		glog.Warningf("Unable to log to stderr by default. Using stdout.")
	}

	var rootCmd = &cobra.Command{
		Use:   "authz --config <config.yaml>",
		Short: "authz service, alpha.",
		Long:  `authz service based on zanzibar access systems. alpha`,
		Run:   serve,
	}

	rootCmd.PersistentFlags().StringP("config", "c", "", "path to config.yaml")

	if err := rootCmd.Execute(); err != nil {
		glog.Fatalf("error running command: %v", err)
	}

}

func serve(cmd *cobra.Command, _ []string) {
	configPath, err := nonEmptyStringFlag("config", cmd.Flags())

	if err != nil {
		_ = cmd.Usage()
		return
	}

	glog.Infof("Starting authz service with config from: %v", configPath)

	go handleSignals()
	bootstrap.Run(configPath)
}

func handleSignals() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for sig := range sigs {
		glog.Infof("Signal received: %s", sig)
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			bootstrap.Stop()
			close(sigs)
		}
	}
}

// nonEmptyStringFlag attempts to get a non-empty string flag from the provided flag set or panic
func nonEmptyStringFlag(flagName string, flags *pflag.FlagSet) (string, error) {
	flagVal := mustGetString(flagName, flags)

	//also check for leading/trailing whitespaces
	if strings.TrimSpace(flagVal) == "" {
		return "", fmt.Errorf(undefinedValueMessage(flagName))
	}
	return flagVal, nil
}

func mustGetString(flagName string, flags *pflag.FlagSet) string {
	flagVal, err := flags.GetString(flagName)
	if err != nil {
		glog.Fatalf(notFoundMessage(flagName, err))
	}
	return flagVal
}

func undefinedValueMessage(flagName string) string {
	return fmt.Sprintf("flag %s needs a defined value.", flagName)
}

func notFoundMessage(flagName string, err error) string {
	return fmt.Sprintf("could not get flag %s from flag set: %s", flagName, err.Error())
}
