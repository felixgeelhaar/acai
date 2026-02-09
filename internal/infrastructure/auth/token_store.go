// Package auth implements the authentication infrastructure.
// This is the adapter for the domain auth.Service port.
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	domain "github.com/felixgeelhaar/acai/internal/domain/auth"
)

// credentialFileVersion is the current credential file format version.
const credentialFileVersion = "1"

// tokenFile is the stored credential format.
type tokenFile struct {
	Version      string    `json:"version"`
	Method       string    `json:"method"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	Workspace    string    `json:"workspace"`
}

// FileTokenStore persists credentials to a JSON file.
type FileTokenStore struct {
	dir string
}

func NewFileTokenStore(dir string) *FileTokenStore {
	return &FileTokenStore{dir: dir}
}

func (s *FileTokenStore) Save(_ context.Context, cred domain.Credential) error {
	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return err
	}

	f := tokenFile{
		Version:      credentialFileVersion,
		Method:       string(cred.Method()),
		AccessToken:  cred.Token().AccessToken(),
		RefreshToken: cred.Token().RefreshToken(),
		ExpiresAt:    cred.Token().ExpiresAt(),
		Workspace:    cred.Workspace(),
	}

	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path(), data, 0600)
}

func (s *FileTokenStore) Load(_ context.Context) (*domain.Credential, error) {
	data, err := os.ReadFile(s.path())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, domain.ErrNotAuthenticated
		}
		return nil, err
	}

	var f tokenFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}

	token := domain.NewToken(f.AccessToken, f.RefreshToken, f.ExpiresAt)
	cred := domain.NewCredential(domain.AuthMethod(f.Method), token, f.Workspace)

	return cred, nil
}

func (s *FileTokenStore) Delete(_ context.Context) error {
	if err := os.Remove(s.path()); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (s *FileTokenStore) path() string {
	return filepath.Join(s.dir, "credentials.json")
}
