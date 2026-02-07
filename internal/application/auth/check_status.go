package auth

import (
	"context"

	domain "github.com/felixgeelhaar/acai/internal/domain/auth"
)

type CheckStatusOutput struct {
	Credential    *domain.Credential
	Authenticated bool
}

type CheckStatus struct {
	service domain.Service
}

func NewCheckStatus(service domain.Service) *CheckStatus {
	return &CheckStatus{service: service}
}

func (uc *CheckStatus) Execute(ctx context.Context) (*CheckStatusOutput, error) {
	cred, err := uc.service.Status(ctx)
	if err != nil {
		return &CheckStatusOutput{Authenticated: false}, nil
	}

	return &CheckStatusOutput{
		Credential:    cred,
		Authenticated: cred.IsValid(),
	}, nil
}
