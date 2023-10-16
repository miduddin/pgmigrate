package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

func loadPgService(name string) (map[string]string, error) {
	pgservicePath := os.Getenv("PGSERVICEFILE")
	if pgservicePath == "" {
		home, _ := os.UserHomeDir()
		pgservicePath = filepath.Join(home, ".pg_service.conf")
	}

	os.Unsetenv("PGSERVICEFILE")

	cfg, err := ini.Load(pgservicePath)
	if err != nil {
		return nil, fmt.Errorf("load pg_service file: %w", err)
	}

	s := cfg.Section(name)
	if s == nil {
		return nil, fmt.Errorf("service not exist")
	}

	return s.KeysHash(), nil
}
