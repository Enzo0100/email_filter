package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Database representa a conexão com o banco de dados
type Database struct {
	pool *pgxpool.Pool
}

// Config configurações do banco de dados
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDatabase cria uma nova instância de conexão com o banco de dados
func NewDatabase(ctx context.Context, cfg *Config) (*Database, error) {
	if cfg == nil {
		cfg = &Config{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvOrDefault("DB_PORT", "5432"),
			User:     getEnvOrDefault("DB_USER", "postgres"),
			Password: getEnvOrDefault("DB_PASSWORD", "postgres"),
			DBName:   getEnvOrDefault("DB_NAME", "email_filter"),
			SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
		}
	}

	// Construir string de conexão
	connString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	// Configurar pool de conexões
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear config do banco: %v", err)
	}

	// Configurações do pool
	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5

	// Criar pool de conexões
	pool, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao banco: %v", err)
	}

	return &Database{pool: pool}, nil
}

// GetConnection retorna uma conexão do pool
func (db *Database) GetConnection(ctx context.Context) (*pgxpool.Conn, error) {
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao adquirir conexão: %v", err)
	}
	return conn, nil
}

// Close fecha o pool de conexões
func (db *Database) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// SetTenantContext adiciona o ID do tenant ao contexto para multitenancy
func SetTenantContext(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, "tenant_id", tenantID)
}

// GetTenantID recupera o ID do tenant do contexto
func GetTenantID(ctx context.Context) string {
	if tenantID, ok := ctx.Value("tenant_id").(string); ok {
		return tenantID
	}
	return ""
}

// ExecuteInTransaction executa uma função dentro de uma transação
func (db *Database) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	conn, err := db.GetConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %v", err)
	}

	if err := fn(ctx, tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("erro no rollback após erro: %v (erro original: %v)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("erro ao commit da transação: %v", err)
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
