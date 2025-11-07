package main

import (
	"fmt"
	"log/slog"

	"github.com/csnewman/team-cli/internal/team"
	"github.com/spf13/cobra"
)

func configureCmdRun(cmd *cobra.Command, args []string) error {
	remoteCfg, err := team.ExtractConfig(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	slog.Info("Extracted remote configuration", "cfg", remoteCfg)

	token, err := team.FetchToken(cmd.Context(), remoteCfg)
	if err != nil {
		return err
	}

	slog.Info("Fetched initial token")

	existingCfg, err := readConfig()
	if err != nil {
		return fmt.Errorf("failed to read existing config: %w", err)
	}

	existingCfg.ServerConfig = remoteCfg
	existingCfg.AuthToken = token

	if err := writeConfig(existingCfg); err != nil {
		return fmt.Errorf("failed to write existing config: %w", err)
	}

	slog.Info("TEAM CLI config updated")

	return nil
}
