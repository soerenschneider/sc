package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/afero"
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
	Run: func(cmd *cobra.Command, args []string) {
		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))

		principals, err := cmd.Flags().GetStringArray(sshSignKeyCmdFlagsPrincipals)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get principals")
		}

		if len(principals) == 0 {
			var suggestions []string
			currentUser, err := user.Current()
			if err == nil {
				suggestions = []string{currentUser.Username}
			}
			principals = []string{tui.ReadInput("Enter principal", suggestions)}
		}

		publicKeyFile, err := cmd.Flags().GetString(sshSignKeyCmdFlagsPublicKeyFile)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get flag")
		}

		publicKeyFile = pkg.GetExpandedFile(publicKeyFile)

		certificateFile, err := cmd.Flags().GetString(sshSignKeyCmdFlagsCertificateFile)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get flag")
		}
		if certificateFile == "" && publicKeyFile != "" {
			certificateFile = strings.Replace(publicKeyFile, ".pub", "", 1)
			certificateFile = pkg.GetExpandedFile(fmt.Sprintf("%s-cert.pub", certificateFile))
		}

		ttl, err := cmd.Flags().GetString(sshSignKeyCmdFlagsTtl)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get flag")
		}

		mount, err := cmd.Flags().GetString(sshCmdFlagsSshMount)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get flag")
		}

		role := pkg.GetString(cmd, sshSignKeyCmdFlagsVaultRole)
		if role == "" {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			availableRoles, err := client.SshListRoles(ctx, mount)
			if err == nil {
				role = tui.SelectInput("Enter role", availableRoles)
			} else {
				role = tui.ReadInput("Enter role", nil)
			}
		}

		fs := afero.NewOsFs()
		publicKeyData, err := afero.ReadFile(fs, publicKeyFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Could not read public key data")
		}

		requestNewCertificate, err := needsRequestNewCertificate(fs, certificateFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Can not proceed")
		}

		if !requestNewCertificate {
			return
		}

		req := vault.SshSignatureRequest{
			Ttl:        ttl,
			Principals: principals,
			Extensions: nil, // TODO
			VaultRole:  role,
			PublicKey:  string(publicKeyData),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		var certificateData string
		if err := spinner.New().
			Type(spinner.Line).
			ActionWithErr(func(ctx context.Context) error {
				certificateData, err = client.SignSshKey(ctx, mount, req)
				return err
			}).
			Title("Sending signature request...").
			Accessible(false).
			Context(ctx).
			Type(spinner.Dots).
			Run(); err != nil {
			log.Fatal().Err(err).Msg("could not sign public key")
		}

		if err := afero.WriteFile(fs, certificateFile, []byte(certificateData), 0640); err != nil {
			log.Error().Err(err).Msg("could not write signature")
		}

		cert, err := vault.ParseSshCertData([]byte(certificateData))
		if err != nil {
			log.Fatal().Err(err).Msg("Could not parse received cert data")
		}
		log.Info().Msgf("New certificate written to disk, valid until %v (%v)", cert.ValidBefore, time.Until(cert.ValidBefore).Round(time.Hour))
	},
}

func needsRequestNewCertificate(fs afero.Fs, certificateFile string) (bool, error) {
	_, err := os.Stat(certificateFile)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		log.Info().Msg("No existing certificate, requesting new certificate")
		return true, nil
	}

	// TODO: maybe need to check more potential errors
	certData, err := afero.ReadFile(fs, certificateFile)
	if err != nil {
		return false, errors.New("could not read certificate file")
	}

	cert, err := vault.ParseSshCertData(certData)
	if err != nil {
		log.Warn().Err(err).Msg("could not parse existing cert data, requesting new certificate")
		return true, nil
	}

	requestNewCertificate := time.Until(cert.ValidBefore) < time.Minute*3 || cert.GetPercentage() < 5
	if time.Now().After(cert.ValidBefore) {
		log.Info().Msgf("Certificate exists but already expired %v ago, requesting new one", time.Since(cert.ValidBefore).Round(time.Minute))
	} else {
		action := "Not requesting new certificate"
		if requestNewCertificate {
			action = "Requesting new certificate"
		}
		log.Info().Msgf("%s, certificate at %.0f%% (expires in %v)", action, cert.GetPercentage(), time.Until(cert.ValidBefore).Round(time.Minute))
	}

	return requestNewCertificate, nil
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

/*
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
*/
