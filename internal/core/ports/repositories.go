package ports

import (
	"context"
	"mi-api-go/internal/core/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByMatricula(ctx context.Context, matricula string) (*domain.User, error)
	GetAll(ctx context.Context) ([]domain.User, error)
	Update(ctx context.Context, user *domain.User) error // 👈 AGREGA ESTA LÍNEA
	Delete(ctx context.Context, id int64) error          // 👈 AGREGA ESTA LÍNEA
}
