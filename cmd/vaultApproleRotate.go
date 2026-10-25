package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sort"
	"time"

	"charm.land/huh/v2/spinner"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/internal/vault/prompt"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	vaultApproleCidr     = "cidr"
	vaultApproleCidrHost = "cidr-host"
	vaultApproleMetadata = "metadata"
	vaultApproleJson     = "json"
	vaultApproleWrapTtl  = "wrap-ttl"
)

type approleRotation struct {
	RoleName         string   `json:"role_name"`
	SecretID         string   `json:"secret_id,omitempty"`
	SecretIDAccessor string   `json:"secret_id_accessor,omitempty"`
	WrappingToken    string   `json:"wrapping_token,omitempty"`
	WrappingAccessor string   `json:"wrapping_accessor,omitempty"`
	WrappingTtl      int      `json:"wrapping_ttl,omitempty"`
	Destroyed        []string `json:"destroyed_accessors"`
	Cidrs            []string `json:"cidrs,omitempty"`
}

var vaultApproleRotateCmd = &cobra.Command{
	Use:     "rotate-secret-id",
	Aliases: []string{"rotate"},
	Short:   "Rotate the secret_id of an AppRole",
	Long: `Create a new secret_id for an AppRole and destroy all of its previous secret_ids afterwards.

The old secret_ids are only destroyed after the new one has been created successfully. If
destroying an old secret_id fails, the new secret_id is printed nonetheless and the command
exits with an error.

The role is given using --role or --role-file. If it is not given and the command is run in a
terminal, the AppRoles are listed for selection.

Using --cidr, the new secret_id and the tokens issued with it are bound to the given CIDR blocks.
Using --cidr-host, the CIDR blocks are looked up dynamically: the host is resolved using DNS and
each of its IP addresses is added as a single host CIDR (/32 for IPv4, /128 for IPv6). Both
flags can be combined.

Using --wrap-ttl, the new secret_id is response wrapped: instead of the secret_id, a single-use
wrapping token valid for the given TTL is printed, which can be unwrapped using
"vault unwrap <token>" or handed to Vault Agent.

When run without a terminal, e.g. by cron, --role (or --role-file) and --force are required.

Use --json to print the new secret_id as JSON, e.g. for further processing with jq.`,
	Example: `  sc vault approle rotate-secret-id
  sc vault approle rotate-secret-id -r my-role
  sc vault approle rotate-secret-id -r my-role --cidr 10.0.0.0/8 --cidr 192.168.1.10/32 -f --json
  sc vault approle rotate-secret-id -r my-role --cidr-host my-host.example.com -f
  sc vault approle rotate-secret-id -r my-role --wrap-ttl 5m -f --json | jq -r .wrapping_token`,
	Run: func(cmd *cobra.Command, args []string) {
		roleName := mustGetApproleRoleName(cmd)
		opts := mustParseApproleSecretIDOptions(cmd)
		cidrs, metadata, ttl := opts.Cidrs, opts.Metadata, opts.Ttl
		wrapTtl := pkg.GetString(cmd, vaultApproleWrapTtl)
		force := pkg.MustGetBool(cmd, vaultForce)
		if !force && !isTerminal(os.Stdin) {
			log.Fatal().Msgf("not running in a terminal, confirm the rotation using --%s", vaultForce)
		}

		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))
		mount := pkg.GetString(cmd, vaultMountPath)

		if roleName == "" {
			roleName = selectApprole(cmd.Context(), client, mount)
		}

		if !force {
			question := fmt.Sprintf("Really rotate the secret_id of role %q? All existing secret_ids will be destroyed.", roleName)
			if tui.SelectInput(question, []string{"no", "yes"}) != "yes" {
				log.Info().Msg("Aborted, secret_id has not been rotated")
				return
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), vaultDefaultTimeout)
		defer cancel()

		rotation := approleRotation{RoleName: roleName, Cidrs: cidrs}
		var destroyErr error
		rotate := func(ctx context.Context) error {
			result, err := client.RotateAppRoleSecretID(ctx, mount, roleName, metadata, cidrs, ttl, wrapTtl)
			if result == nil {
				return err
			}

			rotation.SecretID, rotation.SecretIDAccessor, rotation.Destroyed = result.SecretID, result.SecretIDAccessor, result.Destroyed
			if result.WrapInfo != nil {
				rotation.WrappingToken, rotation.WrappingAccessor, rotation.WrappingTtl = result.WrapInfo.Token, result.WrapInfo.Accessor, result.WrapInfo.TTL
			}
			// The new secret_id exists, report it before failing on old secret_ids that are left.
			destroyErr = err
			return nil
		}

		// The spinner needs a terminal, rotations are also run unattended, e.g. by cron.
		var err error
		//#nosec:G115
		if term.IsTerminal(int(os.Stderr.Fd())) {
			err = spinner.New().
				ActionWithErr(rotate).
				Title("Rotating secret_id...").
				Context(ctx).
				Type(spinner.Dots).
				WithOutput(os.Stderr).
				Run()
		} else {
			err = rotate(ctx)
		}
		if err != nil {
			log.Fatal().Err(err).Msgf("could not rotate secret_id of role %q", roleName)
		}

		if pkg.MustGetBool(cmd, vaultApproleJson) {
			if rotation.Destroyed == nil {
				rotation.Destroyed = []string{}
			}
			printJson(rotation)
		} else {
			output := map[string]any{
				"role_name":           rotation.RoleName,
				"destroyed_accessors": len(rotation.Destroyed),
				"cidrs":               rotation.Cidrs,
			}
			if rotation.WrappingToken != "" {
				output["wrapping_token"] = rotation.WrappingToken
				output["wrapping_accessor"] = rotation.WrappingAccessor
				output["wrapping_ttl"] = time.Duration(rotation.WrappingTtl) * time.Second
			} else {
				output["secret_id"] = rotation.SecretID
				output["secret_id_accessor"] = rotation.SecretIDAccessor
			}
			tui.PrintMapOutput(output)
		}

		if destroyErr != nil {
			log.Fatal().Err(destroyErr).Msg("new secret_id has been created, but not all old secret_ids could be destroyed")
		}
	},
}

// approleSecretIDOptions are the properties of a secret_id to create, as given by the flags added by
// addApproleSecretIDFlags.
type approleSecretIDOptions struct {
	Cidrs    []string
	Metadata map[string]string
	Ttl      string
}

// addApproleSecretIDFlags adds the flags describing a secret_id to create.
func addApproleSecretIDFlags(cmd *cobra.Command) {
	cmd.Flags().StringArray(vaultApproleCidr, nil, "Bind the secret_id and the tokens issued with it to this CIDR block, can be given multiple times")
	cmd.Flags().StringArray(vaultApproleCidrHost, nil, "Resolve this host using DNS and bind the secret_id and its tokens to its IP addresses, can be given multiple times")
	cmd.Flags().StringToString(vaultApproleMetadata, nil, "Metadata to attach to the secret_id, e.g. --metadata host=foo,env=prod")
	cmd.Flags().String(vaultTtl, "", "TTL of the secret_id, defaults to the secret_id_ttl of the role")
	cmd.Flags().Bool(vaultApproleJson, false, "Print the new secret_id as JSON")
}

// mustParseApproleSecretIDOptions validates and returns the values of the flags added by addApproleSecretIDFlags.
// Hosts given by --cidr-host are resolved to CIDRs.
func mustParseApproleSecretIDOptions(cmd *cobra.Command) approleSecretIDOptions {
	cidrs := pkg.GetStringArray(cmd, vaultApproleCidr)
	for _, cidr := range cidrs {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			log.Fatal().Err(err).Msgf("invalid cidr %q", cidr)
		}
	}

	if hosts := pkg.GetStringArray(cmd, vaultApproleCidrHost); len(hosts) > 0 {
		lookupCtx, lookupCancel := context.WithTimeout(context.Background(), vaultDefaultTimeout)
		hostCidrs, err := lookupHostCidrs(lookupCtx, net.DefaultResolver.LookupIPAddr, hosts)
		lookupCancel()
		if err != nil {
			log.Fatal().Err(err).Msg("could not look up cidrs")
		}
		log.Info().Strs("cidrs", hostCidrs).Msgf("Resolved %v", hosts)
		cidrs = append(cidrs, hostCidrs...)
	}

	metadata, err := cmd.Flags().GetStringToString(vaultApproleMetadata)
	if err != nil {
		log.Fatal().Err(err).Msg("could not parse metadata")
	}
	if len(metadata) == 0 {
		metadata = nil
	}

	return approleSecretIDOptions{
		Cidrs:    cidrs,
		Metadata: metadata,
		Ttl:      pkg.GetString(cmd, vaultTtl),
	}
}

// mustGetApproleRoleName returns the role given by --role or --role-file. An empty name is returned if neither is
// given and stdin is a terminal, so the role can be selected interactively using selectApprole.
func mustGetApproleRoleName(cmd *cobra.Command) string {
	roleName, err := readValueOrFile(pkg.GetString(cmd, vaultApproleRoleName), pkg.GetString(cmd, vaultApproleRoleNameFile), "role")
	if err != nil {
		log.Fatal().Err(err).Msg("could not read role name")
	}
	if roleName == "" && !isTerminal(os.Stdin) {
		log.Fatal().Msgf("the role name must be given using --%s or --%s", vaultApproleRoleName, vaultApproleRoleNameFile)
	}
	return roleName
}

// printJson prints v as indented JSON to stdout.
func printJson(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		log.Fatal().Err(err).Msg("could not encode output")
	}
}

// lookupHostCidrs resolves the hosts and returns a single host CIDR for each of their IP addresses.
func lookupHostCidrs(ctx context.Context, lookup func(ctx context.Context, host string) ([]net.IPAddr, error), hosts []string) ([]string, error) {
	var cidrs []string
	seen := map[string]bool{}
	for _, host := range hosts {
		addrs, err := lookup(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("could not resolve %q: %w", host, err)
		}
		if len(addrs) == 0 {
			return nil, fmt.Errorf("no ip addresses found for %q", host)
		}

		for _, addr := range addrs {
			bits := 128
			if addr.IP.To4() != nil {
				bits = 32
			}
			cidr := (&net.IPNet{IP: addr.IP, Mask: net.CIDRMask(bits, bits)}).String()
			if !seen[cidr] {
				seen[cidr] = true
				cidrs = append(cidrs, cidr)
			}
		}
	}
	return cidrs, nil
}

// selectApprole lets the user select one of the AppRoles of the mount. Unlike tui.SelectOrReadInput, a
// selection is offered even if there is only a single role, so the role is always chosen explicitly. Falls
// back to free-form input if the AppRoles can not be listed.
func selectApprole(ctx context.Context, client *vault.VaultClient, mount string) string {
	roles, err := prompt.Load(ctx, "Loading AppRoles...", func(ctx context.Context) ([]string, error) {
		return client.ListAppRoles(ctx, mount)
	})
	if err != nil {
		log.Warn().Err(err).Msg("could not list AppRoles, falling back to manual input")
	}

	if len(roles) == 0 {
		return tui.ReadInput("Enter AppRole", nil)
	}

	sort.Strings(roles)
	return tui.SelectInput("Select AppRole", roles)
}

func init() {
	vaultApproleCmd.AddCommand(vaultApproleRotateCmd)

	addApproleSecretIDFlags(vaultApproleRotateCmd)
	vaultApproleRotateCmd.Flags().BoolP(vaultForce, "f", false, "Skip the confirmation prompt")
	vaultApproleRotateCmd.Flags().String(vaultApproleWrapTtl, "", "Response wrap the new secret_id with the given TTL (e.g. 5m) and print the wrapping token instead")
}
