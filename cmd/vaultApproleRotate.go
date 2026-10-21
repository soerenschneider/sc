package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"charm.land/huh/v2/spinner"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	vaultApproleCidr     = "cidr"
	vaultApproleCidrHost = "cidr-host"
	vaultApproleMetadata = "metadata"
	vaultApproleJson     = "json"
)

type approleRotation struct {
	RoleName         string   `json:"role_name"`
	SecretID         string   `json:"secret_id"`
	SecretIDAccessor string   `json:"secret_id_accessor"`
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

Using --cidr, the new secret_id and the tokens issued with it are bound to the given CIDR blocks.
Using --cidr-host, the CIDR blocks are looked up dynamically: the host is resolved using DNS and
each of its IP addresses is added as a single host CIDR (/32 for IPv4, /128 for IPv6). Both
flags can be combined.

Use --json to print the new secret_id as JSON, e.g. for further processing with jq.`,
	Example: `  sc vault approle rotate-secret-id -r my-role
  sc vault approle rotate-secret-id -r my-role --cidr 10.0.0.0/8 --cidr 192.168.1.10/32 -f --json
  sc vault approle rotate-secret-id -r my-role --cidr-host my-host.example.com -f`,
	Run: func(cmd *cobra.Command, args []string) {
		roleName, err := readValueOrFile(pkg.GetString(cmd, vaultApproleRoleName), pkg.GetString(cmd, vaultApproleRoleNameFile), "role")
		if err != nil {
			log.Fatal().Err(err).Msg("could not read role name")
		}
		if roleName == "" {
			log.Fatal().Msgf("the role name must be given using --%s or --%s", vaultApproleRoleName, vaultApproleRoleNameFile)
		}

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

		if !pkg.MustGetBool(cmd, vaultForce) {
			question := fmt.Sprintf("Really rotate the secret_id of role %q? All existing secret_ids will be destroyed.", roleName)
			if tui.SelectInput(question, []string{"no", "yes"}) != "yes" {
				log.Info().Msg("Aborted, secret_id has not been rotated")
				return
			}
		}

		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))
		mount := pkg.GetString(cmd, vaultMountPath)
		ttl := pkg.GetString(cmd, vaultTtl)

		ctx, cancel := context.WithTimeout(context.Background(), vaultDefaultTimeout)
		defer cancel()

		rotation := approleRotation{RoleName: roleName, Cidrs: cidrs}
		var destroyErr error
		rotate := func(ctx context.Context) error {
			var err error
			rotation.SecretID, rotation.SecretIDAccessor, rotation.Destroyed, err = client.RotateAppRoleSecretID(ctx, mount, roleName, metadata, cidrs, ttl)
			if rotation.SecretID != "" {
				// The new secret_id exists, report it before failing on old secret_ids that are left.
				destroyErr = err
				return nil
			}
			return err
		}

		// The spinner needs a terminal, rotations are also run unattended, e.g. by cron.
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
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(rotation); err != nil {
				log.Fatal().Err(err).Msg("could not encode output")
			}
		} else {
			tui.PrintMapOutput(map[string]any{
				"role_name":           rotation.RoleName,
				"secret_id":           rotation.SecretID,
				"secret_id_accessor":  rotation.SecretIDAccessor,
				"destroyed_accessors": len(rotation.Destroyed),
				"cidrs":               rotation.Cidrs,
			})
		}

		if destroyErr != nil {
			log.Fatal().Err(destroyErr).Msg("new secret_id has been created, but not all old secret_ids could be destroyed")
		}
	},
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

func init() {
	vaultApproleCmd.AddCommand(vaultApproleRotateCmd)

	vaultApproleRotateCmd.Flags().StringArray(vaultApproleCidr, nil, "Bind the secret_id and the tokens issued with it to this CIDR block, can be given multiple times")
	vaultApproleRotateCmd.Flags().StringArray(vaultApproleCidrHost, nil, "Resolve this host using DNS and bind the secret_id and its tokens to its IP addresses, can be given multiple times")
	vaultApproleRotateCmd.Flags().StringToString(vaultApproleMetadata, nil, "Metadata to attach to the secret_id, e.g. --metadata host=foo,env=prod")
	vaultApproleRotateCmd.Flags().String(vaultTtl, "", "TTL of the secret_id, defaults to the secret_id_ttl of the role")
	vaultApproleRotateCmd.Flags().Bool(vaultApproleJson, false, "Print the new secret_id as JSON")
	vaultApproleRotateCmd.Flags().BoolP(vaultForce, "f", false, "Skip the confirmation prompt")
}
