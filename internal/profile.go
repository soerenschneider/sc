package internal

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

type Profile map[string]map[string]map[string]any // profile -> command -> flags

var profileData Profile

func Load(path string) error {
	data, err := os.ReadFile(pkg.GetExpandedFile(path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// file does not exist, ignore it
			return nil
		}
		return err
	}

	if err := yaml.Unmarshal(data, &profileData); err != nil {
		return err
	}

	return nil
}

func getCommandHierarchy(cmd string) []string {
	elements := strings.Split(cmd, "-")
	ret := make([]string, len(elements))

	// Build from longest to shortest by removing suffixes
	current := cmd
	for i := 0; i < len(elements); i++ {
		ret[i] = current

		// Remove the last element and its separator for next iteration
		if i < len(elements)-1 {
			lastIdx := strings.LastIndex(current, "-")
			if lastIdx != -1 {
				current = current[:lastIdx]
			}
		}
	}

	return ret
}

// Get returns the value of key for the given command in the profile, see lookup.
// Lists are joined with commas.
func Get(cmdName, profileName, key string) (string, bool) {
	value, _, ok := lookup(cmdName, profileName, key)
	if !ok {
		return "", false
	}
	values, _, ok := toStrings(value)
	if !ok {
		return "", false
	}
	return strings.Join(values, ","), true
}

// lookup returns the value of key from the most specific section of the
// command hierarchy that sets it, e.g. from "vault-login" before "vault", and
// the name of that section.
func lookup(cmdName, profileName, key string) (any, string, bool) {
	cmd, ok := profileData[profileName]
	if !ok {
		return nil, "", false
	}
	for _, c := range getCommandHierarchy(cmdName) {
		if value, ok := cmd[c][key]; ok {
			return value, c, true
		}
	}
	return nil, "", false
}

// toStrings converts a scalar or a list of scalars from the profile to strings.
// isList reports whether value is a list.
func toStrings(value any) (values []string, isList bool, ok bool) {
	scalar := func(v any) (string, bool) {
		switch v := v.(type) {
		case string:
			return v, true
		case bool:
			return strconv.FormatBool(v), true
		case int:
			return strconv.Itoa(v), true
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64), true
		}
		return "", false
	}

	if list, isList := value.([]any); isList {
		values = make([]string, 0, len(list))
		for _, item := range list {
			v, ok := scalar(item)
			if !ok {
				return nil, true, false
			}
			values = append(values, v)
		}
		return values, true, true
	}
	v, ok := scalar(value)
	return []string{v}, false, ok
}

// Decode decodes the most specific section of the profile for the given command into out, allowing commands to
// read structured settings that can not be expressed as flags. Fields of out that are not present in the section
// keep their value. It returns false if the profile contains no section for the command.
func Decode(cmdName, profileName string, out any) (bool, error) {
	cmd, ok := profileData[profileName]
	if !ok {
		return false, nil
	}

	for _, c := range getCommandHierarchy(cmdName) {
		section, ok := cmd[c]
		if !ok {
			continue
		}

		data, err := yaml.Marshal(section)
		if err != nil {
			return true, err
		}
		if err := yaml.Unmarshal(data, out); err != nil {
			return true, fmt.Errorf("invalid settings for %q in profile %q: %w", c, profileName, err)
		}
		return true, nil
	}

	return false, nil
}

var ErrProfileNotFound = errors.New("profile not found")

// ApplyFlags sets the flags of cmd that were not given on the command line to
// the values of the profile, see lookup.
func ApplyFlags(cmdName string, cmd *cobra.Command, profileName string) error {
	_, found := profileData[profileName]
	if !found {
		return ErrProfileNotFound
	}

	for _, key := range getAllFlagNames(cmd) {
		value, section, ok := lookup(cmdName, profileName, key)
		if !ok {
			continue
		}

		flags := cmd.Flags()
		flag := flags.Lookup(key)
		if flag == nil {
			flags = cmd.PersistentFlags()
			flag = flags.Lookup(key)
		}
		if flag == nil || flags.Changed(key) || flag.Value.String() != flag.DefValue {
			continue
		}

		values, isList, ok := toStrings(value)
		if !ok {
			log.Warn().Msgf("ignoring %q in section %q of profile %q, expected a value or a list of values", key, section, profileName)
			continue
		}

		var err error
		if sliceValue, isSlice := flag.Value.(pflag.SliceValue); isSlice && isList {
			// Set would split the joined values again for slice flags, but
			// not for array flags
			if err = sliceValue.Replace(values); err == nil {
				flag.Changed = true
			}
		} else {
			err = flags.Set(key, strings.Join(values, ","))
		}
		if err != nil {
			log.Warn().Err(err).Msgf("could not set flag %q from section %q of profile %q", key, section, profileName)
			continue
		}
		log.Debug().Msgf("Set flag %q from section %q of profile %q", key, section, profileName)
	}

	return nil
}

// getAllFlagNames returns all local + persistent flag names for a command
func getAllFlagNames(cmd *cobra.Command) []string {
	flagNames := []string{}

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		flagNames = append(flagNames, flag.Name)
	})

	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		// Avoid duplicates if persistent flag also in local flags
		found := false
		for _, name := range flagNames {
			if name == f.Name {
				found = true
				break
			}
		}
		if !found {
			flagNames = append(flagNames, f.Name)
		}
	})

	return flagNames
}
