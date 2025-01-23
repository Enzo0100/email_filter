package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/enzo010/email-filter/internal/domain/entities"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type EmailRepository struct {
	db *Database
}

func NewEmailRepository(db *Database) *EmailRepository {
	return &EmailRepository{db: db}
}

func (r *EmailRepository) Create(ctx context.Context, email *entities.Email) error {
	// Gerar IDs se não fornecidos
	if email.TenantID == "" {
		email.TenantID = uuid.New().String()
	}
	if email.UserID == "" {
		email.UserID = uuid.New().String()
	}

	return r.db.ExecuteInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Inserir email
		query := `
			INSERT INTO emails (
				tenant_id, user_id, subject, from_address, to_address,
				content, priority, category, processed_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id, created_at, updated_at`

		err := tx.QueryRow(
			ctx, query,
			email.TenantID, email.UserID, email.Subject,
			email.From, email.To, email.Content,
			email.Priority, email.Category, email.ProcessedAt,
		).Scan(&email.ID, &email.CreatedAt, &email.UpdatedAt)

		if err != nil {
			return fmt.Errorf("erro ao inserir email: %v", err)
		}

		// Inserir labels
		if len(email.Labels) > 0 {
			for _, label := range email.Labels {
				_, err = tx.Exec(ctx,
					"INSERT INTO email_labels (email_id, label) VALUES ($1, $2)",
					email.ID, label,
				)
				if err != nil {
					return fmt.Errorf("erro ao inserir label: %v", err)
				}
			}
		}

		// Inserir tarefas
		if len(email.Tasks) > 0 {
			for i := range email.Tasks {
				task := &email.Tasks[i]
				query = `
					INSERT INTO tasks (
						email_id, description, due_date,
						priority, status
					) VALUES ($1, $2, $3, $4, $5)
					RETURNING id, created_at, updated_at`

				err = tx.QueryRow(
					ctx, query,
					email.ID, task.Description, task.DueDate,
					task.Priority, task.Status,
				).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)

				if err != nil {
					return fmt.Errorf("erro ao inserir tarefa: %v", err)
				}
			}
		}

		return nil
	})
}

func (r *EmailRepository) GetByID(ctx context.Context, id string) (*entities.Email, error) {
	email := &entities.Email{}

	err := r.db.ExecuteInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Buscar email
		query := `
			SELECT id, tenant_id, user_id, subject, from_address,
				   to_address, content, priority, category,
				   processed_at, created_at, updated_at
			FROM emails WHERE id = $1`

		err := tx.QueryRow(ctx, query, id).Scan(
			&email.ID, &email.TenantID, &email.UserID,
			&email.Subject, &email.From, &email.To,
			&email.Content, &email.Priority, &email.Category,
			&email.ProcessedAt, &email.CreatedAt, &email.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("erro ao buscar email: %v", err)
		}

		// Buscar labels
		rows, err := tx.Query(ctx,
			"SELECT label FROM email_labels WHERE email_id = $1",
			id,
		)
		if err != nil {
			return fmt.Errorf("erro ao buscar labels: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var label string
			if err := rows.Scan(&label); err != nil {
				return fmt.Errorf("erro ao ler label: %v", err)
			}
			email.Labels = append(email.Labels, label)
		}

		// Buscar tarefas
		rows, err = tx.Query(ctx, `
			SELECT id, description, due_date, priority,
				   status, created_at, updated_at
			FROM tasks WHERE email_id = $1`,
			id,
		)
		if err != nil {
			return fmt.Errorf("erro ao buscar tarefas: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var task entities.Task
			if err := rows.Scan(
				&task.ID, &task.Description, &task.DueDate,
				&task.Priority, &task.Status,
				&task.CreatedAt, &task.UpdatedAt,
			); err != nil {
				return fmt.Errorf("erro ao ler tarefa: %v", err)
			}
			email.Tasks = append(email.Tasks, task)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return email, nil
}

func (r *EmailRepository) ListByTenant(ctx context.Context, tenantID string, filters map[string]interface{}) ([]*entities.Email, error) {
	var emails []*entities.Email

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
		WITH filtered_emails AS (
			SELECT id, tenant_id, user_id, subject, from_address,
				   to_address, content, priority, category,
				   processed_at, created_at, updated_at
			FROM emails
			WHERE tenant_id = $1`

	// Adicionar filtros
	args := []interface{}{tenantID}
	argCount := 2

	if category, ok := filters["category"]; ok {
		query += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, category)
		argCount++
	}

	if priority, ok := filters["priority"]; ok {
		query += fmt.Sprintf(" AND priority = $%d", argCount)
		args = append(args, priority)
		argCount++
	}

	if startDate, ok := filters["start_date"]; ok {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, startDate)
		argCount++
	}

	if endDate, ok := filters["end_date"]; ok {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, endDate)
		argCount++
	}

	// Adicionar ordenação e paginação
	query += fmt.Sprintf(`
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	)`, argCount, argCount+1)
	args = append(args, pageSize, offset)

	// Adicionar joins para labels e tasks
	query += `
		SELECT
			e.*,
			ARRAY_AGG(DISTINCT el.label) FILTER (WHERE el.label IS NOT NULL) as labels,
			ARRAY_AGG(DISTINCT jsonb_build_object(
				'id', t.id,
				'description', t.description,
				'due_date', t.due_date,
				'priority', t.priority,
				'status', t.status,
				'created_at', t.created_at,
				'updated_at', t.updated_at
			)) FILTER (WHERE t.id IS NOT NULL) as tasks
		FROM filtered_emails e
		LEFT JOIN email_labels el ON e.id = el.email_id
		LEFT JOIN tasks t ON e.id = t.email_id
		GROUP BY e.id, e.tenant_id, e.user_id, e.subject, e.from_address,
				 e.to_address, e.content, e.priority, e.category,
				 e.processed_at, e.created_at, e.updated_at`

	rows, err := r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar emails: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		email := &entities.Email{}
		var labelsArray []string
		var tasksJson []byte

		err := rows.Scan(
			&email.ID, &email.TenantID, &email.UserID,
			&email.Subject, &email.From, &email.To,
			&email.Content, &email.Priority, &email.Category,
			&email.ProcessedAt, &email.CreatedAt, &email.UpdatedAt,
			&labelsArray, &tasksJson,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler email: %v", err)
		}

		// Processar labels
		if labelsArray != nil {
			email.Labels = labelsArray
		}

		// Processar tasks
		if tasksJson != nil {
			var tasks []entities.Task
			if err := json.Unmarshal(tasksJson, &tasks); err != nil {
				return nil, fmt.Errorf("erro ao decodificar tasks: %v", err)
			}
			email.Tasks = tasks
		}

		emails = append(emails, email)
	}

	return emails, nil
}

func (r *EmailRepository) Update(ctx context.Context, email *entities.Email) error {
	return r.db.ExecuteInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Atualizar email
		query := `
			UPDATE emails SET
				subject = $1,
				from_address = $2,
				to_address = $3,
				content = $4,
				priority = $5,
				category = $6,
				processed_at = $7,
				updated_at = NOW()
			WHERE id = $8 AND tenant_id = $9 AND user_id = $10
			RETURNING updated_at`

		err := tx.QueryRow(
			ctx, query,
			email.Subject, email.From, email.To,
			email.Content, email.Priority, email.Category,
			email.ProcessedAt, email.ID, email.TenantID,
			email.UserID,
		).Scan(&email.UpdatedAt)

		if err != nil {
			return fmt.Errorf("erro ao atualizar email: %v", err)
		}

		// Atualizar labels
		// Primeiro remove todas as labels existentes
		_, err = tx.Exec(ctx,
			"DELETE FROM email_labels WHERE email_id = $1",
			email.ID,
		)
		if err != nil {
			return fmt.Errorf("erro ao remover labels antigas: %v", err)
		}

		// Inserir novas labels
		if len(email.Labels) > 0 {
			for _, label := range email.Labels {
				_, err = tx.Exec(ctx,
					"INSERT INTO email_labels (email_id, label) VALUES ($1, $2)",
					email.ID, label,
				)
				if err != nil {
					return fmt.Errorf("erro ao inserir nova label: %v", err)
				}
			}
		}

		// Atualizar tarefas existentes e adicionar novas
		for i := range email.Tasks {
			task := &email.Tasks[i]
			if task.ID != "" {
				// Atualizar tarefa existente
				query = `
					UPDATE tasks SET
						description = $1,
						due_date = $2,
						priority = $3,
						status = $4,
						updated_at = NOW()
					WHERE id = $5 AND email_id = $6
					RETURNING updated_at`

				err = tx.QueryRow(
					ctx, query,
					task.Description, task.DueDate,
					task.Priority, task.Status,
					task.ID, email.ID,
				).Scan(&task.UpdatedAt)

				if err != nil {
					return fmt.Errorf("erro ao atualizar tarefa: %v", err)
				}
			} else {
				// Inserir nova tarefa
				query = `
					INSERT INTO tasks (
						email_id, description, due_date,
						priority, status
					) VALUES ($1, $2, $3, $4, $5)
					RETURNING id, created_at, updated_at`

				err = tx.QueryRow(
					ctx, query,
					email.ID, task.Description, task.DueDate,
					task.Priority, task.Status,
				).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)

				if err != nil {
					return fmt.Errorf("erro ao inserir nova tarefa: %v", err)
				}
			}
		}

		return nil
	})
}

func (r *EmailRepository) Delete(ctx context.Context, id string) error {
	return r.db.ExecuteInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Deletar labels primeiro devido à chave estrangeira
		_, err := tx.Exec(ctx,
			"DELETE FROM email_labels WHERE email_id = $1",
			id,
		)
		if err != nil {
			return fmt.Errorf("erro ao deletar labels: %v", err)
		}

		// Deletar tarefas devido à chave estrangeira
		_, err = tx.Exec(ctx,
			"DELETE FROM tasks WHERE email_id = $1",
			id,
		)
		if err != nil {
			return fmt.Errorf("erro ao deletar tarefas: %v", err)
		}

		// Deletar o email
		result, err := tx.Exec(ctx,
			"DELETE FROM emails WHERE id = $1",
			id,
		)
		if err != nil {
			return fmt.Errorf("erro ao deletar email: %v", err)
		}

		if result.RowsAffected() == 0 {
			return fmt.Errorf("email não encontrado com id: %s", id)
		}

		return nil
	})
}

func (r *EmailRepository) ListByUser(ctx context.Context, userID string, filters map[string]interface{}) ([]*entities.Email, error) {
	var emails []*entities.Email

	query := `
		SELECT id, tenant_id, user_id, subject, from_address,
			   to_address, content, priority, category,
			   processed_at, created_at, updated_at
		FROM emails 
		WHERE user_id = $1`

	// Adicionar filtros
	args := []interface{}{userID}
	argCount := 2

	if category, ok := filters["category"]; ok {
		query += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, category)
		argCount++
	}

	if priority, ok := filters["priority"]; ok {
		query += fmt.Sprintf(" AND priority = $%d", argCount)
		args = append(args, priority)
		argCount++
	}

	if startDate, ok := filters["start_date"]; ok {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, startDate)
		argCount++
	}

	if endDate, ok := filters["end_date"]; ok {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, endDate)
	}

	query += " ORDER BY created_at DESC LIMIT 100"

	rows, err := r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar emails: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		email := &entities.Email{}
		err := rows.Scan(
			&email.ID, &email.TenantID, &email.UserID,
			&email.Subject, &email.From, &email.To,
			&email.Content, &email.Priority, &email.Category,
			&email.ProcessedAt, &email.CreatedAt, &email.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler email: %v", err)
		}
		emails = append(emails, email)
	}

	return emails, nil
}
