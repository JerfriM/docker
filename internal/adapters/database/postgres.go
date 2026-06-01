package database

import (
	"context"
	"errors"
	"mi-api-go/internal/core/domain"
	"mi-api-go/internal/core/ports"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) ports.UserRepository {
	return &PostgresRepository{pool: pool}
}

// 1. CREATE: Cambiado email por matricula en SQL y en los parámetros
func (r *PostgresRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (name, matricula, password, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
	return r.pool.QueryRow(ctx, query, user.Name, user.Matricula, user.Password, user.CreatedAt).Scan(&user.ID)
}

// 2. RENOMBRADO: De GetByEmail a GetByMatricula, adaptando la query y el Scan
func (r *PostgresRepository) GetByMatricula(ctx context.Context, matricula string) (*domain.User, error) {
	query := `SELECT id, name, matricula, password, created_at FROM users WHERE matricula = $1`
	var u domain.User
	err := r.pool.QueryRow(ctx, query, matricula).Scan(&u.ID, &u.Name, &u.Matricula, &u.Password, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// 3. GET BY ID: Actualizado el SELECT y el Scan para leer matricula
func (r *PostgresRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `SELECT id, name, matricula, password, created_at FROM users WHERE id = $1`
	var u domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(&u.ID, &u.Name, &u.Matricula, &u.Password, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// 4. GET ALL: Actualizado el SELECT y el bucle Scan para leer matricula
func (r *PostgresRepository) GetAll(ctx context.Context) ([]domain.User, error) {
	query := `SELECT id, name, matricula, created_at FROM users`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Matricula, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// 5. UPDATE: Actualizado el script SQL y la propiedad que le pasas
func (r *PostgresRepository) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users SET name = $1, matricula = $2 WHERE id = $3`
	cmd, err := r.pool.Exec(ctx, query, user.Name, user.Matricula, user.ID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("usuario no encontrado")
	}
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	cmd, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("usuario no encontrado")
	}
	return nil
}