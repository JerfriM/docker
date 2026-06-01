package services

import (
	"context"
	"errors"
	"mi-api-go/internal/core/domain"
	"mi-api-go/internal/core/ports"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo      ports.UserRepository
	jwtSecret string
}

func NewUserService(repo ports.UserRepository, jwtSecret string) ports.UserService {
	return &userService{repo: repo, jwtSecret: jwtSecret}
}

// 1. Cambiamos el parámetro 'email' por 'matricula'
func (s *userService) Register(ctx context.Context, name, matricula, password string) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Name:      name,
		Matricula: matricula, // Ahora sí encuentra la variable 'matricula' que viene por parámetro
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// 2. Cambiamos 'email' por 'matricula' en el Login
func (s *userService) Login(ctx context.Context, matricula, password string) (string, error) {
	// Llamamos al repositorio usando el nuevo método que busca por matrícula
	user, err := s.repo.GetByMatricula(ctx, matricula)
	if err != nil {
		return "", errors.New("credenciales inválidas")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("credenciales inválidas")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString([]byte(s.jwtSecret))
}

func (s *userService) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userService) ListUsers(ctx context.Context) ([]domain.User, error) {
	return s.repo.GetAll(ctx)
}
func (s *userService) UpdateUser(ctx context.Context, id int64, name, matricula string) error {
	user := &domain.User{
		ID:        id,
		Name:      name,
		Matricula: matricula,
	}
	return s.repo.Update(ctx, user)
}

func (s *userService) DeleteUser(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}