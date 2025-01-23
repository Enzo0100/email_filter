package database

import (
	"context"
	"fmt"

	"github.com/enzo010/email-filter/internal/domain/entities"
	"github.com/jackc/pgx/v5"
)

type TenantRepository struct {
	db *Database
}

func NewTenantRepository(db *Database) *TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) Create(ctx context.Context, tenant *entities.Tenant) error {
	query := `
		INSERT INTO tenants (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`

	return r.db.ExecuteInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query,
			tenant.ID,
			tenant.Name,
			tenant.CreatedAt,
			tenant.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("error creating tenant: %w", err)
		}
		return nil
	})
}

func (r *TenantRepository) GetByID(ctx context.Context, id string) (*entities.Tenant, error) {
	query := `
		SELECT id, name, plan, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`

	row := r.db.pool.QueryRow(ctx, query, id)
	return scanTenant(row)
}

func scanTenant(row pgx.Row) (*entities.Tenant, error) {
	var t entities.Tenant
	err := row.Scan(
		&t.ID,
		&t.Name,
		&t.Plan,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error scanning tenant: %w", err)
	}
	return &t, nil
}
