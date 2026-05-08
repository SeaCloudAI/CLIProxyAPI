package main

import (
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/store"
)

func TestResolvePostgresStoreConfig_DisabledWithoutDSN(t *testing.T) {
	t.Parallel()

	cfg, enabled := resolvePostgresStoreConfig(func(...string) (string, bool) {
		return "", false
	}, "/workspace", "/writable")
	if enabled {
		t.Fatal("expected postgres store to be disabled without DSN")
	}
	if cfg != (store.PostgresStoreConfig{}) {
		t.Fatalf("expected zero config when disabled, got %#v", cfg)
	}
}

func TestResolvePostgresStoreConfig_UsesDedicatedSchemaAndTables(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"PGSTORE_DSN":          "postgresql://user:pass@127.0.0.1:5432/cliproxy",
		"PGSTORE_SCHEMA":       "cliproxy_test",
		"PGSTORE_CONFIG_TABLE": "cliproxy_config_test",
		"PGSTORE_AUTH_TABLE":   "cliproxy_auth_test",
		"PGSTORE_LOCAL_PATH":   "/state/cliproxy",
	}

	cfg, enabled := resolvePostgresStoreConfig(func(keys ...string) (string, bool) {
		for _, key := range keys {
			if value, ok := env[key]; ok {
				return value, true
			}
		}
		return "", false
	}, "/workspace", "/writable")
	if !enabled {
		t.Fatal("expected postgres store to be enabled when DSN is present")
	}
	if cfg.DSN != env["PGSTORE_DSN"] {
		t.Fatalf("expected DSN %q, got %q", env["PGSTORE_DSN"], cfg.DSN)
	}
	if cfg.Schema != env["PGSTORE_SCHEMA"] {
		t.Fatalf("expected schema %q, got %q", env["PGSTORE_SCHEMA"], cfg.Schema)
	}
	if cfg.ConfigTable != env["PGSTORE_CONFIG_TABLE"] {
		t.Fatalf("expected config table %q, got %q", env["PGSTORE_CONFIG_TABLE"], cfg.ConfigTable)
	}
	if cfg.AuthTable != env["PGSTORE_AUTH_TABLE"] {
		t.Fatalf("expected auth table %q, got %q", env["PGSTORE_AUTH_TABLE"], cfg.AuthTable)
	}
	if cfg.SpoolDir != env["PGSTORE_LOCAL_PATH"] {
		t.Fatalf("expected spool dir %q, got %q", env["PGSTORE_LOCAL_PATH"], cfg.SpoolDir)
	}
}

func TestResolvePostgresStoreConfig_FallsBackToWritableBase(t *testing.T) {
	t.Parallel()

	cfg, enabled := resolvePostgresStoreConfig(func(keys ...string) (string, bool) {
		if len(keys) > 0 && keys[0] == "PGSTORE_DSN" {
			return "postgresql://user:pass@127.0.0.1:5432/cliproxy", true
		}
		return "", false
	}, "/workspace", "/writable")
	if !enabled {
		t.Fatal("expected postgres store to be enabled when DSN is present")
	}
	if cfg.SpoolDir != "/writable" {
		t.Fatalf("expected writable base fallback, got %q", cfg.SpoolDir)
	}
}

func TestResolvePostgresStoreConfig_FallsBackToWorkingDirectory(t *testing.T) {
	t.Parallel()

	cfg, enabled := resolvePostgresStoreConfig(func(keys ...string) (string, bool) {
		if len(keys) > 0 && keys[0] == "PGSTORE_DSN" {
			return "postgresql://user:pass@127.0.0.1:5432/cliproxy", true
		}
		return "", false
	}, "/workspace", "")
	if !enabled {
		t.Fatal("expected postgres store to be enabled when DSN is present")
	}
	if cfg.SpoolDir != "/workspace" {
		t.Fatalf("expected working directory fallback, got %q", cfg.SpoolDir)
	}
}
