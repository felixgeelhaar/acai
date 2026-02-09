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

func (s *Service) Login(ctx context.Context, params domain.LoginParams) (*domain.Credential, error) {
	switch params.Method {
	case domain.AuthAPIToken:
		if params.APIToken == "" {
			return nil, domain.ErrMissingAPIToken
		}
		if len(params.APIToken) < 10 {
			return nil, domain.ErrInvalidAPIToken
		}
		token := domain.NewToken(params.APIToken, "", time.Now().Add(365*24*time.Hour).UTC())
		cred := domain.NewCredential(domain.AuthAPIToken, token, "default")
		if err := s.store.Save(ctx, *cred); err != nil {
			return nil, err
		}
		return cred, nil

	case domain.AuthOAuth:
		return nil, domain.ErrOAuthNotSupported

	default:
		return nil, domain.ErrUnsupportedAuthMethod
	}
}

func (s *Service) Status(ctx context.Context) (*domain.Credential, error) {
	return s.store.Load(ctx)
}

func (s *Service) Logout(ctx context.Context) error {
	return s.store.Delete(ctx)
}
