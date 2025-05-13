package cmd

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/vault"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
)

const (
	vaultLoginUsername = "username"
	vaultLoginOtp      = "otp"
	vaultLoginMfaId    = "mfa-id"
)

// vaultLoginCmd represents the vaultLogin command
var vaultLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with a Vault server using username and password",
	Long: `Authenticate with a HashiCorp Vault server using the "userpass" authentication method.

This command logs you into Vault using a username and password. If two-factor
authentication is enabled, you may also provide a one-time password (OTP).

After successful login, a Vault token is returned. The token is saved to a file.
If the file cannot be written—due to permission issues or missing directories—the token is printed
to stdout as a fallback.`,
	Run: func(cmd *cobra.Command, args []string) {
		username := pkg.GetString(cmd, vaultLoginUsername)
		otp := pkg.GetString(cmd, vaultLoginOtp)
		mfaId := pkg.GetString(cmd, vaultLoginMfaId)
		mount := pkg.GetString(cmd, VaultMountPath)

		client := vault.MustBuildClient(cmd)

		var err error
		if username == "" {
			var suggestions []string
			currentUser, err := user.Current()
			if err == nil {
				suggestions = []string{currentUser.Username}
			}

			username = huhReadInput("Enter username", suggestions)
		}

		var password string
		if password == "" {
			password = huhReadSensitiveInput(fmt.Sprintf("Enter password for %s", username))
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		var vaultSecret *api.Secret
		sendPasswordFunc := func(ctx context.Context) error {
			if mfaId != "" && otp != "" {
				vaultSecret, err = client.LoginSinglePhase(ctx, username, password, mfaId, otp)
				return err
			} else {
				vaultSecret, err = client.SendPassword(ctx, mount, username, password)
				return err
			}
		}

		if err := spinner.New().
			Type(spinner.Line).
			ActionWithErr(sendPasswordFunc).
			Title("Sending login request...").
			Accessible(false).
			Context(ctx).
			Type(spinner.Dots).
			Run(); err != nil {
			log.Fatal().Err(err).Msg("sending login to request to Vault failed")
		}

		if vaultSecret == nil || vaultSecret.Auth == nil {
			log.Fatal().Msg("received empty response from Vault")
		}

		if vaultSecret.Auth.MFARequirement != nil {
			mfaReq, otpId, err := vault.GetVaultMfaRequirement(vaultSecret)
			if err != nil {
				log.Fatal().Err(err).Msg("error retrieving MFA information")
			}

			if otp == "" {
				otp = huhReadInput(fmt.Sprintf("Enter OTP for id %q", otpId), nil)
			}

			ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			sendOtpFunc := func(ctx context.Context) error {
				vaultSecret, err = client.SendOtp(ctx, mfaReq, otp)
				return err
			}

			if err := spinner.New().
				Type(spinner.Line).
				ActionWithErr(sendOtpFunc).
				Title("Sending OTP...").
				Accessible(false).
				Context(ctx).
				Type(spinner.Dots).
				Run(); err != nil {
				log.Fatal().Err(err).Msg("sending OTP to Vault failed")
			}
		}

		tokenFile, err := vault.GetVaultTokenFilePath()
		if err := os.WriteFile(tokenFile, []byte(vaultSecret.Auth.ClientToken), 0600); err != nil {
			var outputHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1F1F1")).Background(lipgloss.Color("#6C50FF")).Bold(true).Padding(0, 1).MarginRight(1).SetString("Vault token")
			fmt.Println(lipgloss.JoinHorizontal(lipgloss.Center, outputHeader.String(), vaultSecret.Auth.ClientToken))

			log.Fatal().Msgf("Failed to save token to %s: %v", tokenFile, err)
		}

		var outputHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1F1F1")).Background(lipgloss.Color("#6C50FF")).Bold(true).Padding(0, 1).MarginRight(1).SetString("Token saved")
		fmt.Println(lipgloss.JoinHorizontal(lipgloss.Center, outputHeader.String(), tokenFile))
	},
}

func init() {
	vaultCmd.AddCommand(vaultLoginCmd)

	vaultLoginCmd.Flags().StringP(vaultLoginUsername, "u", "", "Username for login")
	vaultLoginCmd.Flags().StringP(VaultMountPath, "m", "userpass", "Vault mount for userpass auth engine")
	vaultLoginCmd.Flags().StringP(vaultLoginOtp, "o", "", "OTP value for non-interactive login")
	vaultLoginCmd.Flags().StringP(vaultLoginMfaId, "", "", "MFA ID")
}
