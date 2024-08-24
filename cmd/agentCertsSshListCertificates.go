package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc-agent/pkg/api"
	"github.com/soerenschneider/sc/internal/agent"
	"github.com/soerenschneider/sc/internal/agent/fmt"
	"github.com/spf13/cobra"
)

var agentCertsSshListCertificatesCmd = &cobra.Command{
	Use:   "list-certificates",
	Short: "List all managed SSH certificates",
	Run: func(cmd *cobra.Command, args []string) {
		client := agent.MustBuildApp(cmd)

		user := api.User
		params := &api.SshGetCertificateConfigListParams{
			Type: &user,
		}
		resp, err := client.SshGetCertificateConfigList(context.Background(), params)
		if err != nil {
			log.Fatal().Err(err).Msg("could not send request")
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		data := agent.MustResponse[api.SshManagedCertificatesList](resp)
		fmt.PrintManagedCertificateConfigs(data.Data)
	},
}

func init() {
	agentCertsSshCmd.AddCommand(agentCertsSshListCertificatesCmd)
}
