package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/csnewman/team-cli/internal/team"
)

type AccountCache struct {
	Version  int
	Accounts map[string]*team.Account
}

func cacheAccounts(acc map[string]*team.Account) error {
	enc, err := json.MarshalIndent(&AccountCache{
		Version:  1,
		Accounts: acc,
	}, "", "    ")
	if err != nil {
		return fmt.Errorf("could not marshal: %w", err)
	}

	path, err := configPath("accounts.json")
	if err != nil {
		return fmt.Errorf("could not determine path: %w", err)
	}

	if err := os.WriteFile(path, enc, 0644); err != nil {
		return fmt.Errorf("could not write: %w", err)
	}

	return nil
}

func getAccountsCache() (*AccountCache, bool, error) {
	path, err := configPath("accounts.json")
	if err != nil {
		return nil, false, fmt.Errorf("could not determine path: %w", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		slog.Debug("Could not read account cache", "err", err)

		return nil, false, nil
	}

	var cache *AccountCache

	if err := json.Unmarshal(raw, &cache); err != nil {
		slog.Warn("Could not parse account cache", "err", err)

		return nil, false, nil
	}

	return cache, true, nil
}
