package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/felixgeelhaar/acai/internal/infrastructure/config"
)

func TestWriteAndReadConfigFile_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	want := config.FileConfig{
		DataSource: "local_cache",
		Granola: config.GranolaFileConfig{
			APIURL:    "https://custom.api.com",
			CachePath: "/custom/cache.json",
		},
	}

	if err := config.WriteConfigFile(path, want); err != nil {
		t.Fatalf("WriteConfigFile: %v", err)
	}

	got, err := config.ReadConfigFile(path)
	if err != nil {
		t.Fatalf("ReadConfigFile: %v", err)
	}

	if got.DataSource != want.DataSource {
		t.Errorf("DataSource = %q, want %q", got.DataSource, want.DataSource)
	}
	if got.Granola.APIURL != want.Granola.APIURL {
		t.Errorf("Granola.APIURL = %q, want %q", got.Granola.APIURL, want.Granola.APIURL)
	}
	if got.Granola.CachePath != want.Granola.CachePath {
		t.Errorf("Granola.CachePath = %q, want %q", got.Granola.CachePath, want.Granola.CachePath)
	}
}

func TestReadConfigFile_NonexistentReturnsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.yaml")

	got, err := config.ReadConfigFile(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.DataSource != "" {
		t.Errorf("DataSource = %q, want empty", got.DataSource)
	}
	if got.Granola.APIURL != "" {
		t.Errorf("Granola.APIURL = %q, want empty", got.Granola.APIURL)
	}
}

func TestWriteConfigFile_Permissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := config.WriteConfigFile(path, config.FileConfig{DataSource: "api"}); err != nil {
		t.Fatalf("WriteConfigFile: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
}

func TestWriteConfigFile_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deep", "config.yaml")

	if err := config.WriteConfigFile(path, config.FileConfig{DataSource: "auto"}); err != nil {
		t.Fatalf("WriteConfigFile: %v", err)
	}

	got, err := config.ReadConfigFile(path)
	if err != nil {
		t.Fatalf("ReadConfigFile: %v", err)
	}
	if got.DataSource != "auto" {
		t.Errorf("DataSource = %q, want auto", got.DataSource)
	}
}
