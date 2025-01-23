package database

import (
	"context"
	"fmt"

	"github.com/enzo010/email-filter/internal/domain/entities"
	"github.com/jackc/pgx/v5"
)

type TaskRepository struct {
	db *Database
}

func NewTaskRepository(db *Database) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *entities.Task, emailID string) error {
	return r.db.ExecuteInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Verificar se o email existe
		var exists bool
		err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM emails WHERE id = $1)", emailID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("erro ao verificar email: %v", err)
		}
		if !exists {
			return fmt.Errorf("email não encontrado com id: %s", emailID)
		}

		// Validar status
		if task.Status != "pending" && task.Status != "completed" {
			return fmt.Errorf("status inválido: %s", task.Status)
		}

		query := `
			INSERT INTO tasks (
				email_id, description, due_date,
				priority, status
			) VALUES ($1, $2, $3, $4, $5)
			RETURNING id, created_at, updated_at`

		err = tx.QueryRow(
			ctx, query,
			emailID, task.Description, task.DueDate,
			task.Priority, task.Status,
		).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)

		if err != nil {
			return fmt.Errorf("erro ao criar tarefa: %v", err)
		}

		return nil
	})
}

func (r *TaskRepository) GetByID(ctx context.Context, id string) (*entities.Task, error) {
	task := &entities.Task{}

	err := r.db.ExecuteInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		query := `
			SELECT id, description, due_date,
				   priority, status, created_at, updated_at
			FROM tasks WHERE id = $1`

		err := tx.QueryRow(ctx, query, id).Scan(
			&task.ID, &task.Description, &task.DueDate,
			&task.Priority, &task.Status,
			&task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("erro ao buscar tarefa: %v", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return task, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *entities.Task) error {
	return r.db.ExecuteInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Verificar se a tarefa existe
		var exists bool
		err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM tasks WHERE id = $1)", task.ID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("erro ao verificar tarefa: %v", err)
		}
		if !exists {
			return fmt.Errorf("tarefa não encontrada com id: %s", task.ID)
		}

		// Validar status
		if task.Status != "pending" && task.Status != "completed" {
			return fmt.Errorf("status inválido: %s", task.Status)
		}

		// Validar prioridade
		if task.Priority != entities.PriorityHigh &&
			task.Priority != entities.PriorityMedium &&
			task.Priority != entities.PriorityLow {
			return fmt.Errorf("prioridade inválida: %s", task.Priority)
		}

		query := `
			UPDATE tasks SET
				description = $1,
				due_date = $2,
				priority = $3,
				status = $4,
				updated_at = NOW()
			WHERE id = $5
			RETURNING updated_at`

		err = tx.QueryRow(
			ctx, query,
			task.Description, task.DueDate,
			task.Priority, task.Status,
			task.ID,
		).Scan(&task.UpdatedAt)

		if err != nil {
			return fmt.Errorf("erro ao atualizar tarefa: %v", err)
		}

		return nil
	})
}

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	return r.db.ExecuteInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "DELETE FROM tasks WHERE id = $1", id)
		if err != nil {
			return fmt.Errorf("erro ao deletar tarefa: %v", err)
		}
		return nil
	})
}

func (r *TaskRepository) ListByEmail(ctx context.Context, emailID string, filters map[string]interface{}) ([]*entities.Task, error) {
	var tasks []*entities.Task

	// Extrair parâmetros de paginação
	page := 1
	pageSize := 20
	if p, ok := filters["page"].(int); ok && p > 0 {
		page = p
	}
	if ps, ok := filters["page_size"].(int); ok && ps > 0 && ps <= 100 {
		pageSize = ps
	}
	offset := (page - 1) * pageSize

	// Construir query base
	query := `
		WITH filtered_tasks AS (
			SELECT id, description, due_date,
				   priority, status, created_at, updated_at
			FROM tasks 
			WHERE email_id = $1`

	// Adicionar filtros
	args := []interface{}{emailID}
	argCount := 2

	if priority, ok := filters["priority"]; ok {
		query += fmt.Sprintf(" AND priority = $%d", argCount)
		args = append(args, priority)
		argCount++
	}

	if status, ok := filters["status"]; ok {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	// Adicionar ordenação e paginação
	query += fmt.Sprintf(`
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	)
	SELECT * FROM filtered_tasks`, argCount, argCount+1)

	args = append(args, pageSize, offset)

	rows, err := r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar tarefas: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		task := &entities.Task{}
		err := rows.Scan(
			&task.ID, &task.Description, &task.DueDate,
			&task.Priority, &task.Status,
			&task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler tarefa: %v", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *TaskRepository) ListPendingTasks(ctx context.Context, userID string, filters map[string]interface{}) ([]*entities.Task, error) {
	var tasks []*entities.Task

	// Extrair parâmetros de paginação
	page := 1
	pageSize := 20
	if p, ok := filters["page"].(int); ok && p > 0 {
		page = p
	}
	if ps, ok := filters["page_size"].(int); ok && ps > 0 && ps <= 100 {
		pageSize = ps
	}
	offset := (page - 1) * pageSize

	// Construir query base
	query := `
		WITH filtered_tasks AS (
			SELECT t.id, t.description, t.due_date,
				   t.priority, t.status, t.created_at, t.updated_at
			FROM tasks t
			JOIN emails e ON t.email_id = e.id
			WHERE e.user_id = $1 AND t.status = 'pending'`

	// Adicionar filtros
	args := []interface{}{userID}
	argCount := 2

	if priority, ok := filters["priority"]; ok {
		query += fmt.Sprintf(" AND t.priority = $%d", argCount)
		args = append(args, priority)
		argCount++
	}

	if startDate, ok := filters["start_date"]; ok {
		query += fmt.Sprintf(" AND t.due_date >= $%d", argCount)
		args = append(args, startDate)
		argCount++
	}

	if endDate, ok := filters["end_date"]; ok {
		query += fmt.Sprintf(" AND t.due_date <= $%d", argCount)
		args = append(args, endDate)
		argCount++
	}

	// Adicionar ordenação e paginação
	query += fmt.Sprintf(`
		ORDER BY t.due_date ASC
		LIMIT $%d OFFSET $%d
	)
	SELECT * FROM filtered_tasks`, argCount, argCount+1)

	args = append(args, pageSize, offset)

	rows, err := r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar tarefas pendentes: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		task := &entities.Task{}
		err := rows.Scan(
			&task.ID, &task.Description, &task.DueDate,
			&task.Priority, &task.Status,
			&task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler tarefa: %v", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
