package entities

import (
	"time"
)

// Tenant representa uma organização no sistema multitenancy
type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Plan      string    `json:"plan"` // free, pro, enterprise
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TenantRepository interface para operações com tenants
type TenantRepository interface {
	Create(tenant *Tenant) error
	GetByID(id string) (*Tenant, error)
	Update(tenant *Tenant) error
	Delete(id string) error
	List() ([]*Tenant, error)
}
