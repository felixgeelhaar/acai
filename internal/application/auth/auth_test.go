package auth_test

import (
	"context"
	"testing"
	"time"

	app "github.com/felixgeelhaar/acai/internal/application/auth"
	domain "github.com/felixgeelhaar/acai/internal/domain/auth"
)

type mockAuthService struct {
	credential *domain.Credential
	loginErr   error
	statusErr  error
}

func (m *mockAuthService) Login(_ context.Context, params domain.LoginParams) (*domain.Credential, error) {
	if m.loginErr != nil {
		return nil, m.loginErr
	}
	return m.credential, nil
}

func (m *mockAuthService) Status(_ context.Context) (*domain.Credential, error) {
	if m.statusErr != nil {
		return nil, m.statusErr
	}
	return m.credential, nil
}

func (m *mockAuthService) Logout(_ context.Context) error {
	return nil
}

func TestLogin_Success(t *testing.T) {
	token := domain.NewToken("access", "refresh", time.Now().Add(1*time.Hour).UTC())
	cred := domain.NewCredential(domain.AuthOAuth, token, "ws")
	svc := &mockAuthService{credential: cred}

	uc := app.NewLogin(svc)
	out, err := uc.Execute(context.Background(), app.LoginInput{
		Method:   domain.AuthAPIToken,
		APIToken: "gra_test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Credential.Workspace() != "ws" {
		t.Errorf("got workspace %q", out.Credential.Workspace())
	}
}

func TestLogin_Error(t *testing.T) {
	svc := &mockAuthService{loginErr: domain.ErrInvalidToken}

	uc := app.NewLogin(svc)
	_, err := uc.Execute(context.Background(), app.LoginInput{
		Method:   domain.AuthAPIToken,
		APIToken: "gra_test",
	})
	if err != domain.ErrInvalidToken {
		t.Errorf("got error %v, want %v", err, domain.ErrInvalidToken)
	}
}

func TestCheckStatus_Authenticated(t *testing.T) {
	token := domain.NewToken("access", "refresh", time.Now().Add(1*time.Hour).UTC())
	cred := domain.NewCredential(domain.AuthOAuth, token, "ws")
	svc := &mockAuthService{credential: cred}

	uc := app.NewCheckStatus(svc)
	out, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.Authenticated {
		t.Error("expected authenticated")
	}
}

func TestCheckStatus_NotAuthenticated(t *testing.T) {
	svc := &mockAuthService{statusErr: domain.ErrNotAuthenticated}

	uc := app.NewCheckStatus(svc)
	out, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Authenticated {
		t.Error("expected not authenticated")
	}
}
