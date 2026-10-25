package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

// nameListOutputFormats are the output formats of commands listing names, such as roles.
var nameListOutputFormats = []string{approleOutputSimple, approleOutputJson}

// addNameListOutputFlag adds the --output flag to a command listing names.
func addNameListOutputFlag(cmd *cobra.Command) {
	cmd.Flags().StringP(vaultApproleOutput, "o", approleOutputSimple, "Output format: "+strings.Join(nameListOutputFormats, ", "))
	_ = cmd.RegisterFlagCompletionFunc(vaultApproleOutput, cobra.FixedCompletions(nameListOutputFormats, cobra.ShellCompDirectiveNoFileComp))
}

// mustGetNameListOutput returns the value of the --output flag added by addNameListOutputFlag.
func mustGetNameListOutput(cmd *cobra.Command) string {
	output := pkg.GetString(cmd, vaultApproleOutput)
	if !slices.Contains(nameListOutputFormats, output) {
		log.Fatal().Msgf("invalid output format %q, must be one of %s", output, strings.Join(nameListOutputFormats, ", "))
	}
	return output
}

// mustPrintNames prints the names to stdout, see printNames.
func mustPrintNames(names []string, output string) {
	if err := printNames(os.Stdout, names, output); err != nil {
		log.Fatal().Err(err).Msg("could not print output")
	}
}

// printNames prints the names sorted alphabetically, either one per line or as JSON array.
func printNames(w io.Writer, names []string, output string) error {
	names = slices.Clone(names)
	sort.Strings(names)

	if output == approleOutputJson {
		if names == nil {
			names = []string{}
		}
		return json.NewEncoder(w).Encode(names)
	}

	for _, name := range names {
		if _, err := fmt.Fprintln(w, name); err != nil {
			return err
		}
	}
	return nil
}
