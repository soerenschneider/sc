package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

const (
	vaultAddr      = "address"
	vaultTokenFile = "token-file"
	vaultMountPath = "mount"
	vaultTtl       = "ttl"
	vaultRoleName  = "role"

	vaultPrivateKeyFile  = "key-file"
	vaultCertificateFile = "cert-file"
	vaultCaFile          = "ca-file"
	vaultCertMinLifetime = "min-lifetime"
	vaultCertMinDuration = "min-duration"

	vaultFullChain = "full-chain"

	vaultPublicKeyFile = "pub-key-file"
	vaultPrincipals    = "principals"

	vaultCommonName = "common-name"
	vaultIpSans     = "ip-sans"
	vaultAltNames   = "alt-names"

	vaultAwsDefaultMount = "aws"

	vaultAwsProfile = "aws-profile"

	vaultAwsDefaultTtl = 3600

	vaultAwsDefaultCredentialsFilename = "~/.aws/credentials" //nolint G101
	vaultAwsDefaultProfile             = "default"

	vaulEntityName     = "entity-name"
	vaultLoginUsername = "username"
	vaultLoginOtp      = "otp"
	vaultLoginMfaId    = "mfa-id"

	vaultDefaultTimeout = 7 * time.Second

	vaultIdentityEntityId = "entity-id"
	vaultTotpMethodId     = "method-id"
	vaultForce            = "force"

	vaultIssuer         = "issuer"
	vaultTotpMethodName = "method-name"
	vaultAlgorithm      = "algorithm"
)

var vaultTotpAlgorithms = []string{"SHA1", "SHA256", "SHA512"}

// vaultCmd represents the vault command
var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(vaultCmd)

	vaultCmd.Flags().StringP(vaultAddr, "a", "", "Vault address. If not specified, tries to read env variable VAULT_ADDR.")
	vaultCmd.Flags().StringP(vaultTokenFile, "t", "~/.vault-token", "Vault token file.")
}
