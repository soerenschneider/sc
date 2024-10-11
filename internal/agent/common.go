package agent

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const (
	AgentCmdFlagsServer   = "server"
	AgentCmdFlagsCertFile = "cert-file"
	AgentCmdFlagsKeyFile  = "key-file"
	AgentCmdFlagsCaFile   = "ca-file"
	AgentCmdFlagsVerbose  = "verbose"
)

func MustBuildApp(cmd *cobra.Command) *ScAgentClient {

	server, err := cmd.Flags().GetString(AgentCmdFlagsServer)
	if err != nil {
		log.Fatal().Err(err).Msg("could not get flag")
	}

	certFile, _ := cmd.Flags().GetString(AgentCmdFlagsCertFile)
	keyFile, _ := cmd.Flags().GetString(AgentCmdFlagsKeyFile)
	caFile, _ := cmd.Flags().GetString(AgentCmdFlagsCaFile)

	var opts []HttpClientOption
	if len(certFile) > 0 && len(keyFile) > 0 {
		opts = append(opts, WithTlsClientCert(certFile, keyFile, caFile))
	}

	app, err := BuildApp(server, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("could not build client for sc-agent")
	}
	return app
}

func MustResponse[T any](resp *http.Response) *T {
	if resp == nil {
		log.Fatal().Msg("empty response")
		return nil // unnecessary statement but makes the linter happy
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Fatal().Int("status", resp.StatusCode).Msg("call unsuccessful")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal().Err(err).Msg("could not read from body")
	}

	if len(data) == 0 {
		return nil
	}

	var ret *T
	if err = json.Unmarshal(data, &ret); err != nil {
		log.Fatal().Err(err).Msg("could not unmarshal data")
	}

	return ret
}
