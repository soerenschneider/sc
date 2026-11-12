package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/agent"
	"github.com/spf13/cobra"
)

var agentPackagesUpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrades all packages",
	Run: func(cmd *cobra.Command, args []string) {
		client := agent.MustBuildApp(cmd)

		resp, err := client.PackagesUpgradeRequestsPost(cmd.Context())
		if err != nil {
			log.Fatal().Err(err).Msg("could not send request")
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		_ = agent.MustResponse[any](resp)
	},
}

func init() {
	agentPackagesCmd.AddCommand(agentPackagesUpgradeCmd)
}
