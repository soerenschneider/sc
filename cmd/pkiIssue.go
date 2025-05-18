package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/pki/builder"
	"github.com/soerenschneider/sc/internal/vault"
	pkg2 "github.com/soerenschneider/sc/pkg"
	"github.com/soerenschneider/vault-pki-cli/pkg"
	"github.com/soerenschneider/vault-pki-cli/pkg/pki"
	"github.com/soerenschneider/vault-pki-cli/pkg/renew_strategy"
	"github.com/soerenschneider/vault-pki-cli/pkg/storage/backend"
	"github.com/soerenschneider/vault-pki-cli/pkg/storage/shape"
	"github.com/spf13/cobra"
)

const (
	pkiIssueCmdFlagsCaFile          = "ca-file"
	pkiIssueCmdFlagsPrivateKeyFile  = "key-file"
	pkiIssueCmdFlagsCertificateFile = "cert-file"
	pkiIssueCmdFlagsTtl             = "ttl"
	pkiIssueCmdFlagsCommonName      = "common-name"
	pkiIssueCmdFlagsVaultRole       = "role"
	pkiIssueCmdFlagsIpSans          = "ip-sans"
	pkiIssueCmdFlagsAltNames        = "alt-names"
)

var pkiIssueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Issues a x509 certificate",
	Run: func(cmd *cobra.Command, args []string) {
		commonName, err := cmd.Flags().GetString(pkiIssueCmdFlagsCommonName)
		pkg2.DieOnErr(err, "could not get flag")

		vaultAddress, err := vault.GetVaultAddress(cmd)
		pkg2.DieOnErr(err, "could not get vault address")

		ttl, err := cmd.Flags().GetString(pkiIssueCmdFlagsTtl)
		pkg2.DieOnErr(err, "could not get flag")

		role, err := cmd.Flags().GetString(pkiIssueCmdFlagsVaultRole)
		pkg2.DieOnErr(err, "could not get flag")

		mount, err := cmd.Flags().GetString(pkiCmdFlagsPkiMount)
		pkg2.DieOnErr(err, "could not get flag")

		ipSans, err := cmd.Flags().GetStringArray(pkiIssueCmdFlagsIpSans)
		pkg2.DieOnErr(err, "could not get flag")

		altNames, err := cmd.Flags().GetStringArray(pkiIssueCmdFlagsAltNames)
		pkg2.DieOnErr(err, "could not get flag")

		privateKeyFile, err := cmd.Flags().GetString(pkiIssueCmdFlagsPrivateKeyFile)
		pkg2.DieOnErr(err, "could not get flag")

		certFile, err := cmd.Flags().GetString(pkiIssueCmdFlagsCertificateFile)
		pkg2.DieOnErr(err, "could not get flag")

		caFile, err := cmd.Flags().GetString(pkiIssueCmdFlagsCaFile)
		pkg2.DieOnErr(err, "could not get flag")

		// expand files
		privateKeyFile = pkg2.GetExpandedFile(privateKeyFile)
		certFile = pkg2.GetExpandedFile(certFile)
		caFile = pkg2.GetExpandedFile(caFile)

		pkiService, err := builder.BuildPki(vaultAddress, mount, role)
		pkg2.DieOnErr(err, "could not build pki service")

		issueArgs := pkg.IssueArgs{
			CommonName: commonName,
			Ttl:        ttl,
			IpSans:     ipSans,
			AltNames:   altNames,
		}

		var caStorage, certStorage, privateKeyStorage pki.StorageImplementation
		if len(caFile) > 0 {
			caStorage, err = backend.NewFilesystemStorageFromUri(caFile)
			pkg2.DieOnErr(err, "could not build ca storage")
		}

		if len(certFile) > 0 {
			certStorage, err = backend.NewFilesystemStorageFromUri(certFile)
			pkg2.DieOnErr(err, "could not build cert storage")
		}

		privateKeyStorage, err = backend.NewFilesystemStorageFromUri(privateKeyFile)
		pkg2.DieOnErr(err, "could not build private key storage")

		storageImpl, err := shape.NewKeyPairStorage(certStorage, privateKeyStorage, caStorage)
		pkg2.DieOnErr(err, "could not build storage impl")

		// depending on the hardware of Vault, the available randomness and the key size it's possible that this
		// operating takes quite some time
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()

		result, err := pkiService.Issue(ctx, storageImpl, issueArgs)
		handlePkiIssueResult(result)
		if err != nil {
			log.Err(err).Msg("could not issue certificate")
		}
	},
}

func handlePkiIssueResult(result pkg.IssueResult) {
	if result.Status == pkg.Issued {
		if result.ExistingCert != nil {
			percentage := fmt.Sprintf("%.1f", renew_strategy.GetPercentage(*result.ExistingCert))
			log.Info().Msgf("Existing certificate at %s%% expired or below threshold, valid from %v until %v", percentage, result.ExistingCert.NotBefore.Format(time.RFC3339), result.ExistingCert.NotAfter.Format(time.RFC3339))
		}
		log.Info().Msgf("New certificate valid until %v (%s)", result.IssuedCert.NotAfter.Format(time.RFC3339), time.Until(result.IssuedCert.NotAfter).Round(time.Second))
	} else if result.Status == pkg.Noop {
		percentage := fmt.Sprintf("%.1f", renew_strategy.GetPercentage(*result.ExistingCert))
		log.Info().Msgf("Existing certificate at %s%%, valid until %v (%s)", percentage, result.ExistingCert.NotAfter.Format(time.RFC3339), time.Until(result.ExistingCert.NotAfter).Round(time.Second))
	}
}

func init() {
	pkiCmd.AddCommand(pkiIssueCmd)

	pkiIssueCmd.Flags().StringP(pkiIssueCmdFlagsCommonName, "n", "", "The CN for the certificate")
	err := pkiIssueCmd.MarkFlagRequired(pkiIssueCmdFlagsCommonName)
	pkg2.DieOnErr(err, "could not mark flag required")

	pkiIssueCmd.Flags().StringP(pkiIssueCmdFlagsPrivateKeyFile, "k", "", "File to save the private key to")
	err = pkiIssueCmd.MarkFlagRequired(pkiIssueCmdFlagsPrivateKeyFile)
	pkg2.DieOnErr(err, "could not mark flag required")

	pkiIssueCmd.Flags().StringP(pkiIssueCmdFlagsVaultRole, "r", "", "Vault role")
	err = pkiIssueCmd.MarkFlagRequired(pkiIssueCmdFlagsVaultRole)
	pkg2.DieOnErr(err, "could not mark flag required")

	pkiIssueCmd.Flags().StringP(pkiIssueCmdFlagsCertificateFile, "c", "", "File to save the certificate to")
	pkiIssueCmd.Flags().String(pkiIssueCmdFlagsCaFile, "", "File to save the ca to")
	pkiIssueCmd.Flags().StringP(pkiIssueCmdFlagsTtl, "t", "24h", "TTL of the certificate")

	pkiIssueCmd.Flags().StringArrayP(pkiIssueCmdFlagsIpSans, "s", nil, "IP Sans")
	pkiIssueCmd.Flags().StringArray(pkiIssueCmdFlagsAltNames, nil, "Alternative names")
}
