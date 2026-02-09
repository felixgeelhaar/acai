// Package auth contains the authentication bounded context.
// It models the user's credentials and authentication state
// as domain concepts, independent of any specific OAuth provider.
package auth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotAuthenticated  = errors.New("not authenticated")
	ErrTokenExpired      = errors.New("token has expired")
	ErrInvalidToken      = errors.New("invalid token")
	ErrOAuthNotSupported     = errors.New("OAuth login is not yet supported; use --method api_token with ACAI_GRANOLA_API_TOKEN")
	ErrMissingAPIToken       = errors.New("API token is required: set ACAI_GRANOLA_API_TOKEN environment variable")
	ErrInvalidAPIToken       = errors.New("API token appears invalid: token must be at least 10 characters")
	ErrUnsupportedAuthMethod = errors.New("unsupported authentication method")
)

// AuthMethod represents how the user authenticates.
type AuthMethod string

const (
	AuthOAuth    AuthMethod = "oauth"
	AuthAPIToken AuthMethod = "api_token"
)

// Token is an immutable value object representing an access token.
type Token struct {
	accessToken  string
	refreshToken string
	expiresAt    time.Time
}

func NewToken(accessToken, refreshToken string, expiresAt time.Time) Token {
	return Token{
		accessToken:  accessToken,
		refreshToken: refreshToken,
		expiresAt:    expiresAt,
	}
}

func (t Token) AccessToken() string   { return t.accessToken }
func (t Token) RefreshToken() string  { return t.refreshToken }
func (t Token) ExpiresAt() time.Time  { return t.expiresAt }
func (t Token) IsExpired() bool       { return time.Now().UTC().After(t.expiresAt) }

// Credential is an entity representing the user's stored authentication state.
type Credential struct {
	method    AuthMethod
	token     Token
	workspace string
	createdAt time.Time
}

func NewCredential(method AuthMethod, token Token, workspace string) *Credential {
	return &Credential{
		method:    method,
		token:     token,
		workspace: workspace,
		createdAt: time.Now().UTC(),
	}
}

func (c *Credential) Method() AuthMethod  { return c.method }
func (c *Credential) Token() Token        { return c.token }
func (c *Credential) Workspace() string   { return c.workspace }
func (c *Credential) CreatedAt() time.Time { return c.createdAt }

func (c *Credential) IsValid() bool {
	return c.token.accessToken != "" && !c.token.IsExpired()
}

// LoginParams holds the parameters for a login request.
type LoginParams struct {
	Method   AuthMethod
	APIToken string
}

// Service is the port for authentication operations.
// Implemented in the infrastructure layer.
type Service interface {
	Login(ctx context.Context, params LoginParams) (*Credential, error)
	Status(ctx context.Context) (*Credential, error)
	Logout(ctx context.Context) error
}
