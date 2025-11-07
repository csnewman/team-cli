package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/csnewman/team-cli/internal/team"
)

type Config struct {
	ServerConfig *team.RemoteConfig `json:"server_config"`
	AuthToken    *team.AuthToken    `json:"auth_token"`
}

func configPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user dir: %w", err)
	}

	teamPath := filepath.Join(homeDir, ".config", "team-cli")

	if err := os.MkdirAll(teamPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create team config dir: %w", err)
	}

	return filepath.Join(teamPath, "config.json"), nil
}

func readConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return new(Config), nil
		}

		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config *Config

	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return config, nil
}

func writeConfig(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	enc, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal config file: %w", err)
	}

	if err := os.WriteFile(path, enc, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
