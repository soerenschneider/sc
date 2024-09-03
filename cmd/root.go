package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

const (
	rootCmdFlagsVerbose     = "verbose"
	rootCmdFlagsNoTelemetry = "no-telemetry"
)

var rootCmd = &cobra.Command{
	Use:               "sc",
	Short:             "Universal Command Line Interface for soeren.cloud",
	DisableAutoGenTag: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool(rootCmdFlagsVerbose)
		setupLogLevel(verbose)

		go func() {
			disableTelemetry, err := cmd.Flags().GetBool(rootCmdFlagsNoTelemetry)
			DieOnErr(err, "could not get flag")

			if disableTelemetry || !strings.HasPrefix(internal.BuildVersion, "v") {
				log.Debug().Str("local_version", internal.BuildVersion).Msg("not performing check for update")
				return
			}

			httpClient := retryablehttp.NewClient().StandardClient()
			releaseNotifier, err := internal.NewReleaseNotifier(httpClient, internal.BuildVersion)
			if err != nil {
				log.Warn().Msg("could not build release notifier")
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			releaseNotifier.CheckRelease(ctx)
		}()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolP(rootCmdFlagsVerbose, "v", false, "Print debug logs")
	rootCmd.PersistentFlags().Bool(rootCmdFlagsNoTelemetry, false, "Do not perform check for updated version")
}

func initConfig() {
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func setupLogLevel(debug bool) {
	level := zerolog.InfoLevel
	if debug {
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)
	//#nosec:G115
	if term.IsTerminal(int(os.Stdout.Fd())) {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05",
		})
	}
}
