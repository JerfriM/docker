package ports

import (
	"context"
	"mi-api-go/internal/core/domain"
)

type UserService interface {
	Register(ctx context.Context, name, matricula, password string) (*domain.User, error) // 👈 Cambiado email por matricula
	Login(ctx context.Context, matricula, password string) (string, error)               // 👈 Cambiado email por matricula
	GetUser(ctx context.Context, id int64) (*domain.User, error)
	ListUsers(ctx context.Context) ([]domain.User, error)
	UpdateUser(ctx context.Context, id int64, name, matricula string) error              // 👈 Agregar para el CRUD
	DeleteUser(ctx context.Context, id int64) error                                      // 👈 Agregar para el CRUD
}