package cmd

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/pki"
	"github.com/soerenschneider/sc/internal/pki/format"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/pkg"

	"github.com/spf13/cobra"
)

type pkiStorage interface {
	WriteCert(certData *vault.PkiCertData) error
	ReadCert() (*x509.Certificate, error)
}

type vaultPkiIssueOpts struct {
	MinDuration time.Duration
	MinLifetime float32
	Mount       string
}

var vaultPkiIssueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Issues a x509 certificate",
	Run: func(cmd *cobra.Command, args []string) {
		client := vault.MustAuthenticateClient(vault.MustBuildClient(cmd))

		issueArgs := vault.PkiIssueArgs{
			CommonName: pkg.GetString(cmd, vaultCommonName),
			Ttl:        pkg.GetString(cmd, vaultTtl),
			Role:       pkg.GetString(cmd, vaultRoleName),
			IpSans:     pkg.GetStringArray(cmd, vaultIpSans),
			AltNames:   pkg.GetStringArray(cmd, vaultAltNames),
		}

		minDurationVal := pkg.GetString(cmd, vaultCertMinDuration)
		minDuration, err := time.ParseDuration(minDurationVal)
		if err != nil {
			log.Fatal().Err(err).Msgf("can not parse %q", vaultCertMinDuration)
		}

		opts := vaultPkiIssueOpts{
			MinDuration: minDuration,
			MinLifetime: pkg.GetFloat32(cmd, vaultCertMinLifetime),
			Mount:       pkg.GetString(cmd, vaultMountPath),
		}

		privateKeyFile := pkg.GetExpandedFile(pkg.GetString(cmd, vaultPrivateKeyFile))
		certFile := pkg.GetExpandedFile(pkg.GetString(cmd, vaultCertificateFile))
		caFile := pkg.GetExpandedFile(pkg.GetString(cmd, vaultCaFile))

		var caStorage, certStorage, privateKeyStorage format.StorageImplementation
		if len(caFile) > 0 {
			caStorage, err = pki.NewFilesystemStorageFromUri(caFile)
			pkg.DieOnErr(err, "could not build ca storage")
		}

		if len(certFile) > 0 {
			certStorage, err = pki.NewFilesystemStorageFromUri(certFile)
			pkg.DieOnErr(err, "could not build cert storage")
		}

		privateKeyStorage, err = pki.NewFilesystemStorageFromUri(privateKeyFile)
		pkg.DieOnErr(err, "could not build private key storage")

		storageImpl, err := format.NewKeyPairStorage(certStorage, privateKeyStorage, caStorage)
		pkg.DieOnErr(err, "could not build storage impl")

		issuePkiCert(client, storageImpl, opts, issueArgs)
	},
}

func issuePkiCert(client *vault.VaultClient, storageImpl pkiStorage, opts vaultPkiIssueOpts, issueArgs vault.PkiIssueArgs) {
	existingCert, err := storageImpl.ReadCert()
	if err != nil && !errors.Is(err, pki.ErrNoCertFound) {
		log.Fatal().Err(err).Msg("could not read certificate")
	}

	ctx, cancel := context.WithTimeout(context.Background(), vaultDefaultTimeout)
	defer cancel()
	if !isIssueNewCert(ctx, client, opts, existingCert) {
		return
	}

	// depending on the hardware running Vault, the available randomness and the key size it's possible that this
	// operating takes quite some time
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	var secret *api.Secret
	if err := spinner.New().
		Type(spinner.Line).
		ActionWithErr(func(ctx context.Context) error {
			secret, err = client.PkiIssueCert(ctx, opts.Mount, issueArgs)
			return err
		}).
		Title("Issuing new x509 certificate...").
		Accessible(false).
		Context(ctx).
		Type(spinner.Dots).
		Run(); err != nil {
		log.Fatal().Err(err).Msg("could not issue x509 certificate")
	}

	issuerCertData := &vault.PkiCertData{
		PrivateKey:  []byte(fmt.Sprintf("%s", secret.Data["private_key"])),
		Certificate: []byte(fmt.Sprintf("%s", secret.Data["certificate"])),
		CaData:      []byte(fmt.Sprintf("%s", secret.Data["issuing_ca"])),
	}

	if err := storageImpl.WriteCert(issuerCertData); err != nil {
		log.Fatal().Err(err).Msg("could not write x509 certificate")
	}

	cert, err := pki.ParseCertPem(issuerCertData.Certificate)
	if err != nil {
		log.Fatal().Err(err).Msg("could not parse issued certificate data")
	}
	log.Info().Msgf("New x509 certificate written to disk, valid until %v (%v)", cert.NotAfter, time.Until(cert.NotAfter).Round(time.Hour))
}

func isIssueNewCert(ctx context.Context, client *vault.VaultClient, opts vaultPkiIssueOpts, cert *x509.Certificate) bool {
	var caChainData []byte
	if err := spinner.New().
		Type(spinner.Line).
		ActionWithErr(func(ctx context.Context) error {
			var err error
			caChainData, err = client.PkiGetCaChain(ctx, opts.Mount)
			return err
		}).
		Title("Fetching ca-chain from Vault").
		Accessible(false).
		Context(ctx).
		Type(spinner.Dots).
		Run(); err != nil {
		log.Warn().Err(err).Msg("could not fetch ca-chain")
	}

	caBlock, _ := pem.Decode(caChainData)
	var ca *x509.Certificate
	if caBlock != nil {
		var err error
		ca, err = x509.ParseCertificate(caBlock.Bytes)
		if err != nil {
			log.Warn().Err(err).Msg("could not parse ca-data")
		}
	}

	certValid, err := isCertValid(opts, ca, cert)
	if cert != nil && err == nil && certValid {
		percentage := pki.GetCertRemainingLifetimePercent(*cert)
		log.Info().Msgf("Existing certificate still valid for %v (%0.1f%%), not issuing new certificate", time.Until(cert.NotAfter).Round(time.Minute), percentage)
		return false
	}

	return true
}

func isCertValid(opts vaultPkiIssueOpts, ca, cert *x509.Certificate) (bool, error) {
	if cert == nil {
		return false, errors.New("cert must not be nil")
	}

	if pki.IsCertExpired(*cert) {
		return false, nil
	}

	if ca != nil {
		err := pki.VerifyCertAgainstCa(cert, ca)
		if err != nil && errors.Is(err, pki.ErrValidationFailed) {
			log.Warn().Err(err).Msg("validation against ca failed, cert has been issued to old ca")
			return false, nil
		}
	}

	expiresSoon, err := pki.MustRenewCert(cert, opts.MinDuration, opts.MinLifetime)
	if expiresSoon {
		wanted := fmt.Sprintf("%s / %0.1f%%", opts.MinDuration, opts.MinLifetime)
		got := fmt.Sprintf("%s / %0.1f%%", time.Until(cert.NotAfter).Round(time.Minute), pki.GetCertRemainingLifetimePercent(*cert))
		log.Info().Str("wanted", wanted).Str("got", got).Msg("Certificate lifetime threshold passed, issuing new certificate")
		return false, err
	}
	return true, err
}

func init() {
	vaultPkiCmd.AddCommand(vaultPkiIssueCmd)

	vaultPkiIssueCmd.Flags().StringP(vaultCommonName, "n", "", "The CN for the certificate")
	_ = vaultPkiIssueCmd.MarkFlagRequired(vaultCommonName)

	vaultPkiIssueCmd.Flags().StringP(vaultPrivateKeyFile, "k", "", "File to save the private key to")
	_ = vaultPkiIssueCmd.MarkFlagRequired(vaultPrivateKeyFile)

	vaultPkiIssueCmd.Flags().StringP(vaultRoleName, "r", "", "Vault role")
	_ = vaultPkiIssueCmd.MarkFlagRequired(vaultRoleName)

	vaultPkiIssueCmd.Flags().StringP(vaultCertificateFile, "c", "", "File to save the certificate to")
	vaultPkiIssueCmd.Flags().String(vaultCaFile, "", "File to save the ca to")
	vaultPkiIssueCmd.Flags().StringP(vaultTtl, "t", "24h", "TTL of the certificate")
	vaultPkiIssueCmd.Flags().Float32P(vaultCertMinLifetime, "l", 10., "Minimum percentage of cert lifetime left before forcing issuing a new certificate")
	vaultPkiIssueCmd.Flags().StringP(vaultCertMinDuration, "d", "15m", "Minimum duration of the cert before forcing issuing a new certificate")

	vaultPkiIssueCmd.Flags().StringArrayP(vaultIpSans, "s", nil, "IP Sans")
	vaultPkiIssueCmd.Flags().StringArray(vaultAltNames, nil, "Alternative names")
}
