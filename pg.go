package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

func loadPgService(name string) (map[string]string, error) {
	home, _ := os.UserHomeDir()
	cfg, err := ini.Load(filepath.Join(home, ".pg_service.conf"))
	if err != nil {
		return nil, fmt.Errorf("load pg_service file: %w", err)
	}

	s := cfg.Section(name)
	if s == nil {
		return nil, fmt.Errorf("service not exist")
	}

	return s.KeysHash(), nil
}
