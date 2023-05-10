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

	// Needed to make `glog` believe that the flags have already been parsed, otherwise
	// every log message is prefixed by an error message stating the flags haven't been
	// parsed.
	_ = flag.CommandLine.Parse([]string{})

	// Always log to stderr by default
	if err := flag.Set("logtostderr", "true"); err != nil {
		glog.Warningf("Unable to log to stderr by default. Using stdout.")
	}

	var rootCmd = &cobra.Command{
		Use:   "authz",
		Short: "authz service, alpha.",
		Long:  `authz service based on zanzibar access systems. alpha`,
		Run:   serve,
	}

	rootCmd.Flags().String("endpoint", "", "endpoint")
	rootCmd.Flags().String("token", "", "token")
	rootCmd.Flags().String("store", "stub", "stub or spicedb")
	rootCmd.Flags().String("oidc-discovery", "", "The full OIDC discovery endpoint of the idp (including the /.well-known/openid-configuration portion)")
	rootCmd.Flags().Bool("useTLS", false, "false for no tls (local dev) and true for TLS")
	if err := rootCmd.Execute(); err != nil {
		glog.Fatalf("error running command: %v", err)
	}

}

func serve(cmd *cobra.Command, _ []string) {
	endpoint := mustGetString("endpoint", cmd.Flags())
	token := mustGetString("token", cmd.Flags())
	store := nonEmptyStringFlag("store", cmd.Flags())
	useTLS := mustGetBool("useTLS", cmd.Flags())
	oidcDiscoveryEndpoint := mustGetString("oidc-discovery", cmd.Flags())

	go handleSignals()
	bootstrap.Run(endpoint, oidcDiscoveryEndpoint, token, store, useTLS)
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
func nonEmptyStringFlag(flagName string, flags *pflag.FlagSet) string {
	flagVal := mustGetString(flagName, flags)

	//also check for leading/trailing whitespaces
	if strings.TrimSpace(flagVal) == "" {
		glog.Fatal(undefinedValueMessage(flagName))
	}
	return flagVal
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

func mustGetBool(flagName string, flags *pflag.FlagSet) bool {
	flagVal, err := flags.GetBool(flagName)
	if err != nil {
		glog.Fatalf(notFoundMessage(flagName, err))
	}
	return flagVal
}
