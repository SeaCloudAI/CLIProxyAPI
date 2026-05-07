package managementasset

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShouldSkipBundledManagementAutoUpdate(t *testing.T) {
	originalEnv := os.Getenv("APP_ENV")
	t.Cleanup(func() {
		if err := os.Setenv("APP_ENV", originalEnv); err != nil {
			t.Fatalf("restore APP_ENV: %v", err)
		}
	})

	staticDir := t.TempDir()
	localPath := filepath.Join(staticDir, managementAssetName)
	if err := os.WriteFile(localPath, []byte("panel"), 0o644); err != nil {
		t.Fatalf("write local asset: %v", err)
	}

	if err := os.Setenv("APP_ENV", "develop"); err != nil {
		t.Fatalf("set APP_ENV: %v", err)
	}
	if !shouldSkipBundledManagementAutoUpdate(staticDir) {
		t.Fatal("expected non-prod environment with bundled asset to skip auto-update")
	}

	if err := os.Setenv("APP_ENV", "prod"); err != nil {
		t.Fatalf("set APP_ENV: %v", err)
	}
	if shouldSkipBundledManagementAutoUpdate(staticDir) {
		t.Fatal("expected prod environment to keep auto-update enabled")
	}

	if err := os.Remove(localPath); err != nil {
		t.Fatalf("remove local asset: %v", err)
	}
	if err := os.Setenv("APP_ENV", "develop"); err != nil {
		t.Fatalf("set APP_ENV: %v", err)
	}
	if shouldSkipBundledManagementAutoUpdate(staticDir) {
		t.Fatal("expected missing bundled asset to keep auto-update enabled")
	}
}
