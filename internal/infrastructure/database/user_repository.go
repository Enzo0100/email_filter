package database

import (
	"context"
	"fmt"

	"github.com/enzo010/email-filter/internal/domain/entities"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db *Database
}

func NewUserRepository(db *Database) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO users (id, tenant_id, name, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	return r.db.ExecuteInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query,
			user.ID,
			user.TenantID,
			user.Name,
			user.Email,
			user.PasswordHash,
			user.CreatedAt,
			user.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("error creating user: %w", err)
		}
		return nil
	})
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	query := `
		SELECT id, tenant_id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	row := r.db.pool.QueryRow(ctx, query, id)
	return scanUser(row)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	query := `
		SELECT id, tenant_id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	row := r.db.pool.QueryRow(ctx, query, email)
	return scanUser(row)
}

func scanUser(row pgx.Row) (*entities.User, error) {
	var u entities.User
	err := row.Scan(
		&u.ID,
		&u.TenantID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error scanning user: %w", err)
	}
	return &u, nil
}
