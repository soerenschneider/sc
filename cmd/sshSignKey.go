package cmd

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/ssh"
	"github.com/soerenschneider/sc/internal/ssh/builder"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/pkg"
	"github.com/soerenschneider/vault-ssh-cli/pkg/signature"
	"github.com/spf13/cobra"
)

const (
	sshSignKeyCmdFlagsPublicKeyFile   = "pub-key-file"
	sshSignKeyCmdFlagsCertificateFile = "cert-file"
	sshSignKeyCmdFlagsPrincipals      = "principals"
	sshSignKeyCmdFlagsTtl             = "ttl"
	sshSignKeyCmdFlagsVaultRole       = "role"
)

var sshSignKeyCmd = &cobra.Command{
	Use:   "sign-key",
	Short: "Signs a SSH public key",
	RunE: func(cmd *cobra.Command, args []string) error {
		principals, err := getPrincipals(cmd)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get principals")
		}

		publicKeyFile, err := cmd.Flags().GetString(sshSignKeyCmdFlagsPublicKeyFile)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get flag")
		}

		publicKeyFile = pkg.GetExpandedFile(publicKeyFile)

		publicKeyStorage, err := ssh.NewAferoSink(publicKeyFile)
		if err != nil {
			return err
		}

		certificateFile, err := cmd.Flags().GetString(sshSignKeyCmdFlagsCertificateFile)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get flag")
		}

		certificateStorage, err := ssh.NewAferoSink(sshSignKeyCmdGetCertificateFile(publicKeyFile, certificateFile))
		if err != nil {
			return err
		}

		ttl, err := cmd.Flags().GetString(sshSignKeyCmdFlagsTtl)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get flag")
		}

		address, err := vault.GetVaultAddress(cmd)
		if err != nil {
			log.Fatal().Err(err).Msg("could not sign public key")
		}

		mount, err := cmd.Flags().GetString(sshCmdFlagsSshMount)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get flag")
		}

		role := pkg.GetString(cmd, sshSignKeyCmdFlagsVaultRole)
		if role == "" {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))
			availableRoles, err := client.SshListRoles(ctx, mount)
			if err == nil {
				role = tui.SelectInput("Enter role", availableRoles)
			} else {
				role = tui.ReadInput("Enter role", nil)
			}
		}

		signer, err := builder.BuildSshSigner(address, mount)
		if err != nil {
			log.Error().Err(err).Msg("could not build client to sign pubkey")
		}

		req := signature.SignatureRequest{
			Ttl:        ttl,
			Principals: principals,
			Extensions: nil, // TODO
			VaultRole:  role,
		}

		result, err := signer.SignUserCert(req, publicKeyStorage, certificateStorage)
		if err != nil {
			log.Error().Err(err).Msg("could not sign key")
		}
		logIssueResults(result)
		return err
	},
}

func init() {
	sshCmd.AddCommand(sshSignKeyCmd)

	sshSignKeyCmd.Flags().StringP(sshSignKeyCmdFlagsPublicKeyFile, "p", "", "Location of the public key to sign")
	if err := sshSignKeyCmd.MarkFlagRequired(sshSignKeyCmdFlagsPublicKeyFile); err != nil {
		log.Fatal().Err(err).Msg("could not mark flag required")
	}

	sshSignKeyCmd.Flags().StringP(sshSignKeyCmdFlagsVaultRole, "r", "", "Vault role")
	if err := sshSignKeyCmd.MarkFlagRequired(sshSignKeyCmdFlagsVaultRole); err != nil {
		log.Fatal().Err(err).Msg("could not mark flag required")
	}

	sshSignKeyCmd.Flags().StringP(sshSignKeyCmdFlagsCertificateFile, "c", "", "Where to save the certificate to")
	sshSignKeyCmd.Flags().StringP(sshSignKeyCmdFlagsTtl, "t", "24h", "TTL of the certificate")

	sshSignKeyCmd.Flags().StringArray(sshSignKeyCmdFlagsPrincipals, nil, "Principals")
}

func sshSignKeyCmdGetCertificateFile(publicKeyFile, certificateFile string) string {
	if len(certificateFile) > 0 {
		return certificateFile
	}

	auto := strings.Replace(publicKeyFile, ".pub", "", 1)
	auto = pkg.GetExpandedFile(fmt.Sprintf("%s-cert.pub", auto))
	return auto
}

func logIssueResults(result *signature.IssueResult) {
	if result == nil || signature.Unknown == result.Status {
		log.Warn().Msg("empty/unknown signature result returned")
		return
	}

	switch result.Status {
	case signature.Noop:
		if result.ExistingCert != nil {
			percentage := result.ExistingCert.GetPercentage()
			secondsUntilExpiry := int64(math.Max(0, time.Until(result.ExistingCert.ValidBefore).Seconds()))
			log.Info().Int64("ttl", secondsUntilExpiry).Int("lifetime", int(percentage)).Msgf("Existing certificate at %2.f%% still valid until %v", percentage, result.ExistingCert.ValidBefore)
		}
	case signature.Issued:
		if result.IssuedCert != nil {
			secondsUntilExpiry := int64(time.Until(result.IssuedCert.ValidBefore).Seconds())
			log.Info().Int64("ttl", secondsUntilExpiry).Msgf("Issued new certificate, valid until %s", result.IssuedCert.ValidBefore)
		}
	}
}
