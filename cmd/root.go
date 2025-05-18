package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/cmd/deps"
	"github.com/soerenschneider/sc/internal"
	"github.com/soerenschneider/sc/pkg"
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

		conditionallyLogLatestReleaseInfo(cmd)
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

func conditionallyLogLatestReleaseInfo(cmd *cobra.Command) {
	disableTelemetry := isDisableTelemetry(cmd)
	//nolint G404: no cryptographic randomness required
	if disableTelemetry || rand.New(rand.NewSource(time.Now().Unix())).Float32() > 0.2 || !strings.HasPrefix(internal.BuildVersion, "v") {
		log.Debug().Str("local_version", internal.BuildVersion).Msg("not performing check for update")
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		httpClient := deps.GetHttpClient()
		releaseNotifier, err := internal.NewReleaseNotifier(httpClient, internal.BuildVersion)
		if err != nil {
			log.Warn().Msg("could not build release notifier")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		releaseNotifier.CheckRelease(ctx)
	}()

	wg.Wait()
}

func isDisableTelemetry(cmd *cobra.Command) bool {
	if cmd.Use == "version" || cmd.Use == "help" || cmd.Use == "docs" {
		return true
	}

	disableTelemetry, err := cmd.Flags().GetBool(rootCmdFlagsNoTelemetry)
	pkg.DieOnErr(err, "could not get flag")
	return disableTelemetry
}
