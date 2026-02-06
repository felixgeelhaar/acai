package auth_test

import (
	"testing"
	"time"

	"github.com/felixgeelhaar/granola-mcp/internal/domain/auth"
)

func TestToken_IsExpired(t *testing.T) {
	future := time.Now().Add(1 * time.Hour).UTC()
	token := auth.NewToken("access", "refresh", future)

	if token.IsExpired() {
		t.Error("token with future expiry should not be expired")
	}
	if token.AccessToken() != "access" {
		t.Errorf("got access token %q", token.AccessToken())
	}
	if token.RefreshToken() != "refresh" {
		t.Errorf("got refresh token %q", token.RefreshToken())
	}
}

func TestToken_ExpiredInPast(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour).UTC()
	token := auth.NewToken("access", "refresh", past)

	if !token.IsExpired() {
		t.Error("token with past expiry should be expired")
	}
}

func TestCredential_IsValid(t *testing.T) {
	future := time.Now().Add(1 * time.Hour).UTC()
	token := auth.NewToken("access", "refresh", future)
	cred := auth.NewCredential(auth.AuthOAuth, token, "my-workspace")

	if !cred.IsValid() {
		t.Error("credential with valid token should be valid")
	}
	if cred.Method() != auth.AuthOAuth {
		t.Errorf("got method %q", cred.Method())
	}
	if cred.Workspace() != "my-workspace" {
		t.Errorf("got workspace %q", cred.Workspace())
	}
}

func TestCredential_InvalidWhenExpired(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour).UTC()
	token := auth.NewToken("access", "refresh", past)
	cred := auth.NewCredential(auth.AuthOAuth, token, "ws")

	if cred.IsValid() {
		t.Error("credential with expired token should not be valid")
	}
}

func TestCredential_InvalidWhenEmptyToken(t *testing.T) {
	future := time.Now().Add(1 * time.Hour).UTC()
	token := auth.NewToken("", "refresh", future)
	cred := auth.NewCredential(auth.AuthOAuth, token, "ws")

	if cred.IsValid() {
		t.Error("credential with empty access token should not be valid")
	}
}
