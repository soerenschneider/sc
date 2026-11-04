package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

// approleInfo is the output of the configuration of a single AppRole.
type approleInfo struct {
	Role               string   `json:"role"`
	TokenPolicies      []string `json:"token_policies"`
	TokenTtlSeconds    int64    `json:"token_ttl_seconds"`
	TokenMaxTtlSeconds int64    `json:"token_max_ttl_seconds"`
	TokenBoundCidrs    []string `json:"token_bound_cidrs"`
	SecretIDTtlSeconds int64    `json:"secret_id_ttl_seconds"`
	SecretIDNumUses    int64    `json:"secret_id_num_uses"`
	SecretIDBoundCidrs []string `json:"secret_id_bound_cidrs"`
}

var vaultApproleListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all AppRoles along with their configuration",
	Long: `Lists all AppRoles of the mount, sorted alphabetically, along with their token policies, token TTLs
and secret_id settings.

If some roles can not be read, the remaining ones are printed nonetheless and the command exits
with an error. To list the secret_ids of the AppRoles, use 'list-accessors'.

The output format is chosen with --output:
  pretty  a styled table (default if stdout is a terminal)
  simple  plain, tab-aligned columns without styling, suited for grep and awk (default otherwise)
  json    JSON, e.g. for further processing with jq`,
	Example: `  sc vault approle list
  sc vault approle list -m my-approle-mount
  sc vault approle list -o simple | awk 'NR > 1 { print $1 }'
  sc vault approle list -o json | jq -r '.[] | select(.secret_id_ttl_seconds == 0) | .role'`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		output := mustGetListOutput(cmd, tableOutputFormats)

		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))
		mount := pkg.GetString(cmd, vaultMountPath)

		var result []approleInfo
		var errs error
		err := tui.RunWithSpinner(cmd.Context(), "Loading AppRoles...", func(ctx context.Context) error {
			result, errs = listApproles(ctx, client, mount)
			return nil
		})
		if err != nil {
			log.Fatal().Err(err).Msg("could not list AppRoles")
		}

		switch output {
		case listOutputJson:
			err = printApprolesJson(os.Stdout, result)
		case listOutputSimple:
			err = printApprolesSimple(os.Stdout, result)
		default:
			printApprolesPretty(result)
		}
		if err != nil {
			log.Fatal().Err(err).Msg("could not print output")
		}

		if errs != nil {
			log.Fatal().Err(errs).Msg("could not read all AppRoles")
		}
	},
}

// listApproles reads the configuration of all roles. Roles that can not be read are skipped, the returned error
// contains all errors that occurred.
func listApproles(ctx context.Context, client *vault.VaultClient, mount string) ([]approleInfo, error) {
	listCtx, cancel := context.WithTimeout(ctx, vaultDefaultTimeout)
	roles, err := client.ListAppRoles(listCtx, mount)
	cancel()
	if err != nil {
		return nil, err
	}
	sort.Strings(roles)

	var errs error
	ret := make([]approleInfo, 0, len(roles))
	for _, role := range roles {
		roleCtx, cancel := context.WithTimeout(ctx, vaultDefaultTimeout)
		approle, err := client.GetAppRole(roleCtx, mount, role)
		cancel()
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
		ret = append(ret, toApproleInfo(approle))
	}

	return ret, errs
}

func toApproleInfo(r *vault.AppRole) approleInfo {
	ret := approleInfo{
		Role:               r.Name,
		TokenPolicies:      r.TokenPolicies,
		TokenTtlSeconds:    int64(r.TokenTtl.Seconds()),
		TokenMaxTtlSeconds: int64(r.TokenMaxTtl.Seconds()),
		TokenBoundCidrs:    r.TokenBoundCidrs,
		SecretIDTtlSeconds: int64(r.SecretIDTtl.Seconds()),
		SecretIDNumUses:    r.SecretIDNumUses,
		SecretIDBoundCidrs: r.SecretIDBoundCidrs,
	}

	if ret.TokenPolicies == nil {
		ret.TokenPolicies = []string{}
	}
	if ret.TokenBoundCidrs == nil {
		ret.TokenBoundCidrs = []string{}
	}
	if ret.SecretIDBoundCidrs == nil {
		ret.SecretIDBoundCidrs = []string{}
	}

	return ret
}

func printApprolesJson(w io.Writer, roles []approleInfo) error {
	if roles == nil {
		roles = []approleInfo{}
	}
	return printListJson(w, roles)
}

func printApprolesSimple(w io.Writer, roles []approleInfo) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "ROLE\tPOLICIES\tTOKEN_TTL\tTOKEN_MAX_TTL\tSECRET_ID_TTL\tSECRET_ID_USES\tSECRET_ID_CIDRS")
	for _, row := range approleRows(roles, listSepSimple) {
		_, _ = fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	return tw.Flush()
}

func printApprolesPretty(roles []approleInfo) {
	nonExpiring := 0
	for _, role := range roles {
		if role.SecretIDTtlSeconds == 0 {
			nonExpiring++
		}
	}

	rows := approleRows(roles, listSepPretty)
	for _, row := range rows {
		if row[4] == "never" {
			row[4] = tui.WarnStyle.Render(row[4])
		}
	}

	tui.Table{
		Title:        "AppRoles",
		Headers:      []string{"Role", "Policies", "Token TTL", "Token max TTL", "Secret-ID TTL", "Secret-ID uses", "Secret-ID CIDRs"},
		Rows:         rows,
		Aligns:       []tui.Align{0, 0, tui.AlignRight, tui.AlignRight, tui.AlignRight, tui.AlignRight},
		Caption:      fmt.Sprintf("%d roles, %d with non-expiring secret_ids", len(roles), nonExpiring),
		Zebra:        len(rows) > 10,
		FitWidth:     true,
		Wrap:         true,
		EmptyMessage: "(no AppRoles)",
	}.Print()
}

// approleRows returns a row per role. A TTL of 0 means the token TTL defaults to the mount's or system's TTL,
// while secret_ids without TTL never expire. Lists within a cell are joined with sep.
func approleRows(roles []approleInfo, sep string) [][]string {
	rows := make([][]string, 0, len(roles))
	for _, role := range roles {
		uses := "unlimited"
		if role.SecretIDNumUses > 0 {
			uses = fmt.Sprintf("%d", role.SecretIDNumUses)
		}

		rows = append(rows, []string{
			role.Role,
			joinOrDash(role.TokenPolicies, sep),
			formatApproleTtl(role.TokenTtlSeconds, "default"),
			formatApproleTtl(role.TokenMaxTtlSeconds, "default"),
			formatApproleTtl(role.SecretIDTtlSeconds, "never"),
			uses,
			joinOrDash(role.SecretIDBoundCidrs, sep),
		})
	}
	return rows
}

// formatApproleTtl formats a TTL in seconds compactly, e.g. "1d12h" instead of "36h0m0s".
func formatApproleTtl(seconds int64, fallback string) string {
	if seconds <= 0 {
		return fallback
	}

	d := time.Duration(seconds) * time.Second
	units := []struct {
		suffix string
		unit   time.Duration
	}{
		{"d", 24 * time.Hour},
		{"h", time.Hour},
		{"m", time.Minute},
		{"s", time.Second},
	}

	var sb strings.Builder
	for _, u := range units {
		if n := d / u.unit; n > 0 {
			_, _ = fmt.Fprintf(&sb, "%d%s", n, u.suffix)
			d -= n * u.unit
		}
	}
	return sb.String()
}

func init() {
	vaultApproleCmd.AddCommand(vaultApproleListCmd)
	addListOutputFlag(vaultApproleListCmd, tableOutputFormats)
}
