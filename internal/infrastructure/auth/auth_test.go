package auth_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	domain "github.com/felixgeelhaar/acai/internal/domain/auth"
	infraauth "github.com/felixgeelhaar/acai/internal/infrastructure/auth"
)

func TestFileTokenStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	store := infraauth.NewFileTokenStore(dir)

	cred := testCredential()
	if err := store.Save(context.Background(), *cred); err != nil {
		t.Fatalf("save error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "credentials.json")); err != nil {
		t.Fatalf("credentials file not found: %v", err)
	}

	loaded, err := store.Load(context.Background())
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.Token().AccessToken() != cred.Token().AccessToken() {
		t.Errorf("got access token %q", loaded.Token().AccessToken())
	}
	if loaded.Workspace() != "test-ws" {
		t.Errorf("got workspace %q", loaded.Workspace())
	}
}

func TestFileTokenStore_LoadNotFound(t *testing.T) {
	dir := t.TempDir()
	store := infraauth.NewFileTokenStore(dir)

	_, err := store.Load(context.Background())
	if err != domain.ErrNotAuthenticated {
		t.Errorf("got error %v, want %v", err, domain.ErrNotAuthenticated)
	}
}

func TestFileTokenStore_Delete(t *testing.T) {
	dir := t.TempDir()
	store := infraauth.NewFileTokenStore(dir)

	cred := testCredential()
	_ = store.Save(context.Background(), *cred)

	if err := store.Delete(context.Background()); err != nil {
		t.Fatalf("delete error: %v", err)
	}

	_, err := store.Load(context.Background())
	if err != domain.ErrNotAuthenticated {
		t.Errorf("got error %v after delete", err)
	}
}

func TestFileTokenStore_DeleteNonexistent(t *testing.T) {
	dir := t.TempDir()
	store := infraauth.NewFileTokenStore(dir)

	if err := store.Delete(context.Background()); err != nil {
		t.Errorf("delete nonexistent should not error: %v", err)
	}
}

func TestService_LoginAndStatus(t *testing.T) {
	dir := t.TempDir()
	store := infraauth.NewFileTokenStore(dir)
	svc := infraauth.NewService(store)

	cred, err := svc.Login(context.Background(), domain.AuthOAuth)
	if err != nil {
		t.Fatalf("login error: %v", err)
	}
	if cred.Method() != domain.AuthOAuth {
		t.Errorf("got method %q", cred.Method())
	}

	status, err := svc.Status(context.Background())
	if err != nil {
		t.Fatalf("status error: %v", err)
	}
	if !status.IsValid() {
		t.Error("credential should be valid after login")
	}
}

func TestService_Logout(t *testing.T) {
	dir := t.TempDir()
	store := infraauth.NewFileTokenStore(dir)
	svc := infraauth.NewService(store)

	_, _ = svc.Login(context.Background(), domain.AuthOAuth)
	if err := svc.Logout(context.Background()); err != nil {
		t.Fatalf("logout error: %v", err)
	}

	_, err := svc.Status(context.Background())
	if err != domain.ErrNotAuthenticated {
		t.Errorf("expected not authenticated after logout, got: %v", err)
	}
}

func testCredential() *domain.Credential {
	token := domain.NewToken("test-access", "test-refresh", time.Now().Add(1*time.Hour).UTC())
	return domain.NewCredential(domain.AuthOAuth, token, "test-ws")
}
