package auth

import (
	"context"
	"time"

	domain "github.com/felixgeelhaar/acai/internal/domain/auth"
)

// TokenStore abstracts token persistence.
type TokenStore interface {
	Save(ctx context.Context, cred domain.Credential) error
	Load(ctx context.Context) (*domain.Credential, error)
	Delete(ctx context.Context) error
}

// Service implements domain.Service for authentication.
type Service struct {
	store TokenStore
}

func NewService(store TokenStore) *Service {
	return &Service{store: store}
}

func (s *Service) Login(ctx context.Context, method domain.AuthMethod) (*domain.Credential, error) {
	// For API token auth, the token is pre-configured.
	// For OAuth, a browser flow would be triggered here.
	// This is a placeholder — the real OAuth flow will be added in Phase 1.
	token := domain.NewToken("placeholder", "", time.Now().Add(24*time.Hour).UTC())
	cred := domain.NewCredential(method, token, "default")

	if err := s.store.Save(ctx, *cred); err != nil {
		return nil, err
	}

	return cred, nil
}

func (s *Service) Status(ctx context.Context) (*domain.Credential, error) {
	return s.store.Load(ctx)
}

func (s *Service) Logout(ctx context.Context) error {
	return s.store.Delete(ctx)
}
