package auth

import (
	"context"

	domain "github.com/felixgeelhaar/acai/internal/domain/auth"
)

type Logout struct {
	service domain.Service
}

func NewLogout(service domain.Service) *Logout {
	return &Logout{service: service}
}

func (uc *Logout) Execute(ctx context.Context) error {
	return uc.service.Logout(ctx)
}
