# Variáveis
APP_NAME=email-filter
MAIN_PATH=cmd/api/main.go
MIGRATION_PATH=internal/infrastructure/database/migrations

# Comandos Go
.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	@go build -o bin/$(APP_NAME) $(MAIN_PATH)

.PHONY: run
run:
	@echo "Running $(APP_NAME)..."
	@go run $(MAIN_PATH)

.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html

.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Docker commands
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(APP_NAME) .

.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	@docker compose up -d

.PHONY: docker-stop
docker-stop:
	@echo "Stopping Docker container..."
	@docker compose down

.PHONY: docker-logs
docker-logs:
	@echo "Showing Docker logs..."
	@docker compose logs -f

# Database commands
.PHONY: db-create
db-create:
	@echo "Creating database..."
	@createdb $(APP_NAME)

.PHONY: db-drop
db-drop:
	@echo "Dropping database..."
	@dropdb $(APP_NAME)

.PHONY: db-migrate
db-migrate:
	@echo "Running database migrations..."
	@psql -d $(APP_NAME) -f $(MIGRATION_PATH)/001_initial_schema.sql

.PHONY: db-reset
db-reset: db-drop db-create db-migrate

# Linting e formatação
.PHONY: lint
lint:
	@echo "Running linter..."
	@golangci-lint run

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Limpeza
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean -testcache

# Ajuda
.PHONY: help
help:
	@echo "Comandos disponíveis:"
	@echo "  make build         - Compila o projeto"
	@echo "  make run          - Executa o servidor"
	@echo "  make test         - Executa os testes"
	@echo "  make deps         - Baixa as dependências"
	@echo "  make docker-build - Constrói a imagem Docker"
	@echo "  make docker-run   - Inicia os containers"
	@echo "  make docker-stop  - Para os containers"
	@echo "  make db-create    - Cria o banco de dados"
	@echo "  make db-migrate   - Executa as migrações"
	@echo "  make db-reset     - Reseta o banco de dados"
	@echo "  make lint         - Executa o linter"
	@echo "  make fmt          - Formata o código"
	@echo "  make clean        - Limpa arquivos temporários"

# Comando padrão
.DEFAULT_GOAL := help