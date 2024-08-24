package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc-agent/pkg/api"
	"github.com/soerenschneider/sc/internal/agent"
	"github.com/spf13/cobra"
)

var agentPowerStateRebootCmd = &cobra.Command{
	Use:   "reboot",
	Short: "Reboots the machine",
	Run: func(cmd *cobra.Command, args []string) {
		client := agent.MustBuildApp(cmd)

		params := &api.PowerPostActionParams{
			Action: api.Reboot,
		}
		resp, err := client.PowerPostAction(context.Background(), params)
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
	agentPowerStateCmd.AddCommand(agentPowerStateRebootCmd)
}
