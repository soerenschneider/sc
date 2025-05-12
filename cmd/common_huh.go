package cmd

import (
	"github.com/charmbracelet/huh"
)

func readNormal(prompt string) (string, error) {
	var input string
	err := huh.NewInput().
		Title(prompt).
		Prompt("> ").
		Value(&input).
		Run()
	if err != nil {
		return "", err
	}
	return input, nil
}

func readSensitive(prompt string) (string, error) {
	var input string
	err := huh.NewInput().
		Title(prompt).
		Prompt("> ").
		EchoMode(huh.EchoModePassword).
		Value(&input).
		Run()
	if err != nil {
		return "", err
	}
	return input, nil
}
