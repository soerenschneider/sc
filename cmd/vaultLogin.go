package cmd

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	vault "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		vaultAddr := cmp.Or(GetString(cmd, vaultAddr), getVaultAddr())
		username := GetString(cmd, vaultLoginUsername)
		otp := GetString(cmd, vaultLoginOtp)
		mfaId := GetString(cmd, vaultLoginMfaId)

		config := vault.DefaultConfig()
		config.Address = vaultAddr

		client, err := vault.NewClient(config)
		if err != nil {
			log.Fatal().Msgf("Error creating Vault client: %v", err)
		}

		if username == "" {
			username, err = readNormal("Enter username")
			if err != nil {
				log.Fatal().Err(err).Msg("could not read username")
			}
		}

		var password string
		if password == "" {
			password, err = readSensitive("Enter password")
			if err != nil {
				log.Fatal().Err(err).Msg("could not read password")
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		var vaultSecret *vault.Secret
		sendPasswordFunc := func(ctx context.Context) error {
			if mfaId != "" && otp != "" {
				vaultSecret, err = vaultLoginSinglePhase(ctx, client, username, password, mfaId, otp)
				return err
			} else {
				vaultSecret, err = vaultSendPassword(ctx, client, username, password)
				return err
			}
		}

		err = spinner.New().
			Type(spinner.Line).
			ActionWithErr(sendPasswordFunc).
			Title("Sending request...").
			Accessible(false).
			Context(ctx).
			Type(spinner.Dots).
			Run()

		if err != nil {
			log.Fatal().Err(err).Msg("sending login to request to Vault failed")
		}

		if vaultSecret == nil || vaultSecret.Auth == nil {
			log.Fatal().Msg("received empty response from Vault")
		}

		if vaultSecret.Auth.MFARequirement != nil {
			mfaReq, otpId, err := getVaultMfaRequirement(vaultSecret)
			if err != nil {
				log.Fatal().Err(err).Msg("yo")
			}

			if otp == "" {
				otp, err = readNormal(fmt.Sprintf("Enter TOTP for id %q", otpId))
				if err != nil {
					log.Fatal().Err(err).Msg("could not read otp")
				}
			}

			ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			sendOtpFunc := func(ctx context.Context) error {
				vaultSecret, err = vaultSendOtp(ctx, client, mfaReq, otp)
				return err
			}

			err = spinner.New().
				Type(spinner.Line).
				ActionWithErr(sendOtpFunc).
				Title("Sending OTP...").
				Accessible(false).
				Context(ctx).
				Type(spinner.Dots).
				Run()

			if err != nil {
				log.Fatal().Err(err).Msg("sending OTP to Vault failed")
			}
		}

		tokenFile, err := getVaultTokenFilePath()
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
	vaultLoginCmd.Flags().StringP(vaultLoginOtp, "o", "", "OTP value for non-interactive login")
	vaultLoginCmd.Flags().StringP(vaultLoginMfaId, "", "", "MFA ID")
}
