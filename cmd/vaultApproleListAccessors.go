package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

const (
	vaultApproleOutput = "output"

	approleOutputPretty = "pretty"
	approleOutputSimple = "simple"
	approleOutputJson   = "json"

	approleStatusValid       = "valid"
	approleStatusExpired     = "expired"
	approleStatusNonExpiring = "non-expiring"

	approleTimeFormat = "2006-01-02 15:04:05"
)

var approleOutputFormats = []string{approleOutputPretty, approleOutputSimple, approleOutputJson}

// approleSecretIDs is the output of the secret_ids of a single AppRole.
type approleSecretIDs struct {
	Role      string            `json:"role"`
	SecretIDs []approleSecretID `json:"secret_ids"`
}

type approleSecretID struct {
	Accessor        string            `json:"accessor"`
	Status          string            `json:"status"`
	CreationTime    *time.Time        `json:"creation_time,omitempty"`
	ExpirationTime  *time.Time        `json:"expiration_time,omitempty"`
	LastUpdatedTime *time.Time        `json:"last_updated_time,omitempty"`
	TtlSeconds      int64             `json:"ttl_seconds"`
	NumUses         int64             `json:"num_uses"`
	CidrList        []string          `json:"cidr_list"`
	TokenBoundCidrs []string          `json:"token_bound_cidrs"`
	Metadata        map[string]string `json:"metadata"`
}

var vaultApproleListAccessorsCmd = &cobra.Command{
	Use:     "list-accessors [role]",
	Aliases: []string{"accessors"},
	Short:   "List the secret_id accessors of AppRoles",
	Long: `Lists the secret_id accessors of all AppRoles, or of a single AppRole if it is passed as argument
or via --role / --role-file, along with whether the secret_ids are expired.

Roles without any secret_id are listed as well. If some roles or accessors can not be read,
the remaining ones are printed nonetheless and the command exits with an error.

The output format is chosen with --output:
  pretty  a styled table (default)
  simple  plain, tab-aligned columns without styling, suited for grep and awk
  json    JSON, e.g. for further processing with jq`,
	Example: `  sc vault approle list-accessors
  sc vault approle list-accessors my-role
  sc vault approle accessors -o simple | grep expired
  sc vault approle list-accessors -o json | jq '.[] | select(.secret_ids | length == 0) | .role'`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		output := pkg.GetString(cmd, vaultApproleOutput)
		if !slices.Contains(approleOutputFormats, output) {
			log.Fatal().Msgf("invalid output format %q, must be one of %s", output, strings.Join(approleOutputFormats, ", "))
		}

		roleName, err := readValueOrFile(pkg.GetString(cmd, vaultApproleRoleName), pkg.GetString(cmd, vaultApproleRoleNameFile), "role")
		if err != nil {
			log.Fatal().Err(err).Msg("could not read role name")
		}
		if len(args) > 0 {
			roleName = args[0]
		}

		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))
		mount := pkg.GetString(cmd, vaultMountPath)

		var result []approleSecretIDs
		var errs error
		err = tui.RunWithSpinner(cmd.Context(), "Loading AppRole secret_ids...", func(ctx context.Context) error {
			result, errs = listApproleSecretIDs(ctx, client, mount, roleName)
			return nil
		})
		if err != nil {
			log.Fatal().Err(err).Msg("could not list AppRole secret_ids")
		}

		switch output {
		case approleOutputJson:
			err = printApproleSecretIDsJson(os.Stdout, result)
		case approleOutputSimple:
			err = printApproleSecretIDsSimple(os.Stdout, result)
		default:
			printApproleSecretIDsPretty(result)
		}
		if err != nil {
			log.Fatal().Err(err).Msg("could not print output")
		}

		if errs != nil {
			log.Fatal().Err(errs).Msg("could not read all AppRole secret_ids")
		}
	},
}

// listApproleSecretIDs looks up the secret_ids of the given role, or of all roles if roleName is empty. Roles and
// accessors that can not be read are skipped, the returned error contains all errors that occurred.
func listApproleSecretIDs(ctx context.Context, client *vault.VaultClient, mount, roleName string) ([]approleSecretIDs, error) {
	roles := []string{roleName}
	if roleName == "" {
		listCtx, cancel := context.WithTimeout(ctx, vaultDefaultTimeout)
		var err error
		roles, err = client.ListAppRoles(listCtx, mount)
		cancel()
		if err != nil {
			return nil, err
		}
		sort.Strings(roles)
	}

	now := time.Now()
	var errs error
	ret := make([]approleSecretIDs, 0, len(roles))
	for _, role := range roles {
		roleCtx, cancel := context.WithTimeout(ctx, vaultDefaultTimeout)
		accessors, err := client.GetAppRoleAccessors(roleCtx, mount, role)
		cancel()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("role %q: %w", role, err))
			continue
		}

		entry := approleSecretIDs{Role: role, SecretIDs: []approleSecretID{}}
		for _, accessor := range accessors {
			lookupCtx, cancel := context.WithTimeout(ctx, vaultDefaultTimeout)
			secretID, err := client.LookupAppRoleSecretIDAccessor(lookupCtx, mount, role, accessor)
			cancel()
			if err != nil {
				errs = multierror.Append(errs, err)
				continue
			}
			entry.SecretIDs = append(entry.SecretIDs, toApproleSecretID(secretID, now))
		}

		// sort by expiration, secret_ids expiring first are listed first and non-expiring ones last
		sort.SliceStable(entry.SecretIDs, func(i, j int) bool {
			a, b := entry.SecretIDs[i].ExpirationTime, entry.SecretIDs[j].ExpirationTime
			if a == nil || b == nil {
				return b == nil && a != nil
			}
			return a.Before(*b)
		})
		ret = append(ret, entry)
	}

	return ret, errs
}

func toApproleSecretID(s *vault.AppRoleSecretID, now time.Time) approleSecretID {
	ret := approleSecretID{
		Accessor:        s.Accessor,
		Status:          approleStatusValid,
		CreationTime:    timeOrNil(s.CreationTime),
		ExpirationTime:  timeOrNil(s.ExpirationTime),
		LastUpdatedTime: timeOrNil(s.LastUpdatedTime),
		TtlSeconds:      int64(s.Ttl.Seconds()),
		NumUses:         s.NumUses,
		CidrList:        s.CidrList,
		TokenBoundCidrs: s.TokenBoundCidrs,
		Metadata:        s.Metadata,
	}

	switch {
	case !s.Expires():
		ret.Status = approleStatusNonExpiring
	case s.Expired(now):
		ret.Status = approleStatusExpired
	}

	if ret.CidrList == nil {
		ret.CidrList = []string{}
	}
	if ret.TokenBoundCidrs == nil {
		ret.TokenBoundCidrs = []string{}
	}
	if ret.Metadata == nil {
		ret.Metadata = map[string]string{}
	}

	return ret
}

func timeOrNil(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func printApproleSecretIDsJson(w io.Writer, roles []approleSecretIDs) error {
	if roles == nil {
		roles = []approleSecretIDs{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(roles)
}

func printApproleSecretIDsSimple(w io.Writer, roles []approleSecretIDs) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "ROLE\tACCESSOR\tSTATUS\tCREATED\tEXPIRES\tUSES\tCIDRS\tMETADATA")
	for _, row := range approleSecretIDRows(roles) {
		_, _ = fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	return tw.Flush()
}

func printApproleSecretIDsPretty(roles []approleSecretIDs) {
	rows := approleSecretIDRows(roles)

	expired := 0
	for _, role := range roles {
		for _, secretID := range role.SecretIDs {
			if secretID.Status == approleStatusExpired {
				expired++
			}
		}
	}

	statusStyles := map[string]lipgloss.Style{
		approleStatusValid:       tui.StatusStyle,
		approleStatusExpired:     tui.ErrorStyle,
		approleStatusNonExpiring: tui.WarnStyle,
	}
	for _, row := range rows {
		if style, ok := statusStyles[row[2]]; ok {
			row[2] = style.Render(row[2])
		}
	}

	tui.Table{
		Title:        "AppRole secret_ids",
		Headers:      []string{"Role", "Accessor", "Status", "Created", "Expires", "Uses", "CIDRs", "Metadata"},
		Rows:         rows,
		Aligns:       []tui.Align{0, 0, 0, 0, 0, tui.AlignRight},
		Caption:      fmt.Sprintf("%d roles, %d secret_ids, %d expired", len(roles), len(rows)-rolesWithoutSecretIDs(roles), expired),
		Zebra:        len(rows) > 10,
		EmptyMessage: "(no AppRoles)",
	}.Print()
}

func rolesWithoutSecretIDs(roles []approleSecretIDs) int {
	ret := 0
	for _, role := range roles {
		if len(role.SecretIDs) == 0 {
			ret++
		}
	}
	return ret
}

// approleSecretIDRows returns a row per secret_id. Roles without secret_id are represented by a row without
// accessor.
func approleSecretIDRows(roles []approleSecretIDs) [][]string {
	var rows [][]string
	for _, role := range roles {
		if len(role.SecretIDs) == 0 {
			rows = append(rows, []string{role.Role, "-", "-", "-", "-", "-", "-", "-"})
			continue
		}

		for _, s := range role.SecretIDs {
			uses := "unlimited"
			if s.NumUses > 0 {
				uses = fmt.Sprintf("%d", s.NumUses)
			}

			rows = append(rows, []string{
				role.Role,
				s.Accessor,
				s.Status,
				formatApproleTime(s.CreationTime, "-"),
				formatApproleTime(s.ExpirationTime, "never"),
				uses,
				joinOrDash(s.CidrList),
				formatApproleMetadata(s.Metadata),
			})
		}
	}
	return rows
}

func formatApproleTime(t *time.Time, fallback string) string {
	if t == nil {
		return fallback
	}
	return t.Local().Format(approleTimeFormat)
}

func formatApproleMetadata(metadata map[string]string) string {
	if len(metadata) == 0 {
		return "-"
	}
	pairs := make([]string, 0, len(metadata))
	for k, v := range metadata {
		pairs = append(pairs, k+"="+v)
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ",")
}

func joinOrDash(items []string) string {
	if len(items) == 0 {
		return "-"
	}
	return strings.Join(items, ",")
}

func init() {
	vaultApproleCmd.AddCommand(vaultApproleListAccessorsCmd)

	vaultApproleListAccessorsCmd.Flags().StringP(vaultApproleOutput, "o", approleOutputPretty, "Output format: "+strings.Join(approleOutputFormats, ", "))
	_ = vaultApproleListAccessorsCmd.RegisterFlagCompletionFunc(vaultApproleOutput, cobra.FixedCompletions(approleOutputFormats, cobra.ShellCompDirectiveNoFileComp))
}
