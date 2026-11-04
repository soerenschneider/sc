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

const (
	vaultListOutput = "output"

	listOutputPretty = "pretty"
	listOutputSimple = "simple"
	listOutputJson   = "json"

	// listSepSimple joins lists within a cell of simple output, without spaces so columns stay awk-friendly.
	listSepSimple = ","
	// listSepPretty joins lists within a cell of pretty output, with spaces so wrapped cells break between items.
	listSepPretty = ", "
)

// tableOutputFormats are the output formats of commands listing tabular data.
var tableOutputFormats = []string{listOutputPretty, listOutputSimple, listOutputJson}

// nameListOutputFormats are the output formats of commands listing names, such as roles.
var nameListOutputFormats = []string{listOutputSimple, listOutputJson}

// addListOutputFlag adds the --output flag to a list command supporting the given formats. If the flag is not
// set, the format is chosen by mustGetListOutput.
func addListOutputFlag(cmd *cobra.Command, formats []string) {
	usage := "Output format: " + strings.Join(formats, ", ")
	if slices.Contains(formats, listOutputPretty) {
		usage += " (default pretty if stdout is a terminal, simple otherwise)"
	} else {
		usage += ` (default "simple")`
	}
	cmd.Flags().StringP(vaultListOutput, "o", "", usage)
	_ = cmd.RegisterFlagCompletionFunc(vaultListOutput, cobra.FixedCompletions(formats, cobra.ShellCompDirectiveNoFileComp))
}

// mustGetListOutput returns the value of the --output flag added by addListOutputFlag. If it is not set, pretty is
// returned if it is supported and stdout is a terminal, simple otherwise.
func mustGetListOutput(cmd *cobra.Command, formats []string) string {
	output := pkg.GetString(cmd, vaultListOutput)
	if output == "" {
		return defaultListOutput(formats, isTerminal(os.Stdout))
	}
	if !slices.Contains(formats, output) {
		log.Fatal().Msgf("invalid output format %q, must be one of %s", output, strings.Join(formats, ", "))
	}
	return output
}

func defaultListOutput(formats []string, terminal bool) string {
	if terminal && slices.Contains(formats, listOutputPretty) {
		return listOutputPretty
	}
	return listOutputSimple
}

// printListJson prints v as indented JSON.
func printListJson(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// addNameListOutputFlag adds the --output flag to a command listing names.
func addNameListOutputFlag(cmd *cobra.Command) {
	addListOutputFlag(cmd, nameListOutputFormats)
}

// mustGetNameListOutput returns the value of the --output flag added by addNameListOutputFlag.
func mustGetNameListOutput(cmd *cobra.Command) string {
	return mustGetListOutput(cmd, nameListOutputFormats)
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

	if output == listOutputJson {
		if names == nil {
			names = []string{}
		}
		return printListJson(w, names)
	}

	for _, name := range names {
		if _, err := fmt.Fprintln(w, name); err != nil {
			return err
		}
	}
	return nil
}
