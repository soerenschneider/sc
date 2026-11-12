package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc-agent/pkg/api"
	"github.com/soerenschneider/sc/internal/agent"
	"github.com/spf13/cobra"
)

const agentServiceStopCmdFlagUnit = "unit"

var agentServiceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops a service unit",
	Run: func(cmd *cobra.Command, args []string) {
		client := agent.MustBuildApp(cmd)

		unit, err := cmd.Flags().GetString(agentServiceStopCmdFlagUnit)
		if err != nil {
			log.Fatal().Err(err).Msg("could not get flag")
		}

		params := &api.ServicesUnitStatusPutParams{
			Action: api.ServicesUnitStatusPutParamsActionStop,
		}
		resp, err := client.ServicesUnitStatusPut(cmd.Context(), unit, params)
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
	agentServiceCmd.AddCommand(agentServiceStopCmd)
	agentServiceStopCmd.Flags().String(agentServiceStopCmdFlagUnit, "", "The unit to stop")
	if err := agentServiceStopCmd.MarkFlagRequired(agentServiceStopCmdFlagUnit); err != nil {
		log.Fatal().Err(err).Msg("could not mark flag required")
	}
}
