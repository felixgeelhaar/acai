package auth

import (
	"context"

	domain "github.com/felixgeelhaar/acai/internal/domain/auth"
)

type LoginInput struct {
	Method   domain.AuthMethod
	APIToken string
}

type LoginOutput struct {
	Credential *domain.Credential
}

type Login struct {
	service domain.Service
}

func NewLogin(service domain.Service) *Login {
	return &Login{service: service}
}

func (uc *Login) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	cred, err := uc.service.Login(ctx, domain.LoginParams{
		Method:   input.Method,
		APIToken: input.APIToken,
	})
	if err != nil {
		return nil, err
	}
	return &LoginOutput{Credential: cred}, nil
}
