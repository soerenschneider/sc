package cmd

import (
	"errors"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/rs/zerolog/log"
)

func huhSelectInput(title string, choices []string) string {
	var input string
	err := huh.NewSelect[string]().
		Title(title).
		Options(huh.NewOptions(choices...)...).
		Value(&input).
		Run()

	if err != nil {
		log.Fatal().Err(err).Msg("could not get choice")
	}

	return input
}

func huhReadInput(title string, suggestions []string) string {
	var input string
	err := huh.NewInput().
		Title(title).
		Suggestions(suggestions).
		Prompt("> ").
		Value(&input).
		Validate(func(val string) error {
			if strings.TrimSpace(val) == "" {
				return errors.New("input can not be empty")
			}
			return nil
		}).
		Run()

	if err != nil {
		log.Fatal().Err(err).Msg("could not read input")
	}

	return input
}

func huhReadSensitiveInput(title string) string {
	var input string
	err := huh.NewInput().
		Title(title).
		Prompt("> ").
		EchoMode(huh.EchoModePassword).
		Value(&input).
		Validate(func(val string) error {
			if strings.TrimSpace(val) == "" {
				return errors.New("input can not be empty")
			}
			return nil
		}).
		Run()

	if err != nil {
		log.Fatal().Err(err).Msg("could not read input")
	}

	return input
}
