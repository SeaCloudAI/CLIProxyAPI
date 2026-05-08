package main

import "github.com/router-for-me/CLIProxyAPI/v6/internal/store"

func resolvePostgresStoreConfig(lookupEnv func(...string) (string, bool), workingDir, writableBase string) (store.PostgresStoreConfig, bool) {
	if lookupEnv == nil {
		return store.PostgresStoreConfig{}, false
	}

	dsn, ok := lookupEnv("PGSTORE_DSN", "pgstore_dsn")
	if !ok {
		return store.PostgresStoreConfig{}, false
	}

	cfg := store.PostgresStoreConfig{
		DSN: dsn,
	}
	if value, ok := lookupEnv("PGSTORE_SCHEMA", "pgstore_schema"); ok {
		cfg.Schema = value
	}
	if value, ok := lookupEnv("PGSTORE_CONFIG_TABLE", "pgstore_config_table"); ok {
		cfg.ConfigTable = value
	}
	if value, ok := lookupEnv("PGSTORE_AUTH_TABLE", "pgstore_auth_table"); ok {
		cfg.AuthTable = value
	}
	if value, ok := lookupEnv("PGSTORE_USAGE_TABLE", "pgstore_usage_table"); ok {
		cfg.UsageTable = value
	}
	if value, ok := lookupEnv("PGSTORE_LOCAL_PATH", "pgstore_local_path"); ok {
		cfg.SpoolDir = value
	}
	if cfg.SpoolDir == "" {
		if writableBase != "" {
			cfg.SpoolDir = writableBase
		} else {
			cfg.SpoolDir = workingDir
		}
	}
	return cfg, true
}
