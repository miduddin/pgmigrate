package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

type pgservice struct {
	Host       string `ini:"host"`
	DBName     string `ini:"dbname"`
	User       string `ini:"user"`
	SearchPath string `ini:"search_path"`
}

func pgConnect(serviceName string) (*sql.DB, error) {
	s, err := loadPgservice(serviceName)
	if err != nil {
		return nil, fmt.Errorf("load pgservice: %w", err)
	}

	// lib/pq does not support PGSERVICEFILE.
	os.Unsetenv("PGSERVICEFILE")
	db, err := sql.Open("postgres", fmt.Sprintf(
		"user=%s host=%s dbname=%s search_path=%s sslmode=disable",
		s.User, s.Host, s.DBName, s.SearchPath,
	))
	if err != nil {
		return nil, fmt.Errorf("sql open: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	fmt.Printf(
		"Connected to host=%s db=%s schema=%s user=%s\n",
		yellow(s.Host), yellow(s.DBName), yellow(s.SearchPath), yellow(s.User),
	)

	return db, nil
}

func loadPgservice(name string) (pgservice, error) {
	cfg, err := ini.Load(pgservicePath())
	if err != nil {
		return pgservice{}, fmt.Errorf("load pg_service file: %w", err)
	}

	s := cfg.Section(name)
	if s == nil {
		return pgservice{}, fmt.Errorf("service not exist")
	}

	ret := pgservice{SearchPath: "public"}
	if err := s.MapTo(&ret); err != nil {
		return pgservice{}, fmt.Errorf("parse pg_service: %w", err)
	}

	return ret, nil
}

func pgservicePath() string {
	if ret := os.Getenv("PGSERVICEFILE"); ret != "" {
		return ret
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".pg_service.conf")
}
