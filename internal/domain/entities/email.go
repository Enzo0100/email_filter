package entities

import (
	"context"
	"time"
)

// Priority representa o nível de prioridade do email
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

// Email representa um email classificado no sistema
type Email struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	UserID      string    `json:"user_id"`
	Subject     string    `json:"subject"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	Content     string    `json:"content"`
	Priority    Priority  `json:"priority"`
	Category    string    `json:"category"`
	Labels      []string  `json:"labels"`
	Tasks       []Task    `json:"tasks"`
	ProcessedAt time.Time `json:"processed_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Task representa uma tarefa sugerida baseada no conteúdo do email
type Task struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	Priority    Priority  `json:"priority"`
	Status      string    `json:"status"` // pending, completed
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"` // Adicionado campo UpdatedAt
}

// EmailRepository interface para operações com emails
type EmailRepository interface {
	Create(ctx context.Context, email *Email) error
	GetByID(ctx context.Context, id string) (*Email, error)
	Update(ctx context.Context, email *Email) error
	Delete(ctx context.Context, id string) error
	ListByTenant(ctx context.Context, tenantID string, filters map[string]interface{}) ([]*Email, error)
	ListByUser(ctx context.Context, userID string, filters map[string]interface{}) ([]*Email, error)
}

// TaskRepository interface para operações com tarefas
type TaskRepository interface {
	Create(task *Task) error
	GetByID(id string) (*Task, error)
	Update(task *Task) error
	Delete(id string) error
	ListByEmail(emailID string, filters map[string]interface{}) ([]*Task, error)
	ListPendingTasks(userID string, filters map[string]interface{}) ([]*Task, error)
}
