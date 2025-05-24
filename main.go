package main

import (
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/cmd"
	"github.com/soerenschneider/sc/internal"
)

func main() {
	err := internal.Load("~/.sc-profiles.yaml")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to load profile data")
	}
	cmd.Execute()
}
