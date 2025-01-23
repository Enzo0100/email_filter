package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/enzo010/email-filter/internal/application/services"
	"github.com/enzo010/email-filter/internal/domain/entities"
	"github.com/enzo010/email-filter/internal/infrastructure/database"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Server struct {
	db              *database.Database
	emailRepo       *database.EmailRepository
	taskRepo        *database.TaskRepository
	userRepo        *database.UserRepository
	tenantRepo      *database.TenantRepository
	emailClassifier *services.EmailClassifier
	emailProcessor  *services.EmailProcessor
	router          *mux.Router
}

func NewServer() (*Server, error) {
	// Carregar variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Printf("Arquivo .env não encontrado: %v", err)
	}

	// Inicializar banco de dados
	db, err := database.NewDatabase(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	// Inicializar repositórios
	emailRepo := database.NewEmailRepository(db)
	taskRepo := database.NewTaskRepository(db)
	userRepo := database.NewUserRepository(db)
	tenantRepo := database.NewTenantRepository(db)

	// Inicializar classificador
	emailClassifier := services.NewEmailClassifier()

	// Configurar processador de email
	emailConfig := &services.EmailConfig{
		Server:   os.Getenv("EMAIL_SERVER"),
		Port:     mustParseInt(os.Getenv("EMAIL_PORT")),
		Username: os.Getenv("EMAIL_USERNAME"),
		Password: os.Getenv("EMAIL_PASSWORD"),
		Folder:   "INBOX",
		SSL:      true,
		TenantID: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", // ID do tenant do script SQL
		UserID:   "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", // ID do usuário do script SQL
	}

	emailProcessor, err := services.NewEmailProcessor(emailConfig, emailClassifier, emailRepo)
	if err != nil {
		return nil, err
	}

	// Inicializar router
	router := mux.NewRouter()

	return &Server{
		db:              db,
		emailRepo:       emailRepo,
		taskRepo:        taskRepo,
		userRepo:        userRepo,
		tenantRepo:      tenantRepo,
		emailClassifier: emailClassifier,
		emailProcessor:  emailProcessor,
		router:          router,
	}, nil
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "300")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) setupRoutes() {
	// Aplicar middleware CORS globalmente
	s.router.Use(corsMiddleware)

	// Endpoints de saúde e métricas
	s.router.HandleFunc("/health", s.healthCheck).Methods("GET")
	s.router.HandleFunc("/metrics", s.metrics).Methods("GET")

	// API v1
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Endpoints de autenticação
	api.HandleFunc("/auth/login", s.handleLogin).Methods("POST", "OPTIONS")

	// Endpoints de emails
	api.HandleFunc("/emails", s.handleClassifyEmail).Methods("POST")
	api.HandleFunc("/emails", s.handleListEmails).Methods("GET")
	api.HandleFunc("/emails/{id}", s.handleGetEmail).Methods("GET")

	// Endpoints de tarefas
	api.HandleFunc("/tasks", s.handleListTasks).Methods("GET")
	api.HandleFunc("/tasks/{id}", s.handleUpdateTask).Methods("PUT")
	api.HandleFunc("/tasks/{id}", s.handleDeleteTask).Methods("DELETE")
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Buscar usuário pelo email
	user, err := s.userRepo.GetByEmail(r.Context(), credentials.Email)
	if err != nil {
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	// TODO: Implementar verificação de senha com hash
	if credentials.Password != "test123" {
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	// Buscar tenant do usuário
	tenant, err := s.tenantRepo.GetByID(r.Context(), user.TenantID)
	if err != nil {
		http.Error(w, "Erro ao buscar tenant", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"user": map[string]interface{}{
			"id":        user.ID,
			"email":     user.Email,
			"name":      user.Name,
			"role":      user.Role,
			"tenantId":  user.TenantID,
			"createdAt": user.CreatedAt,
		},
		"tenant": map[string]interface{}{
			"id":        tenant.ID,
			"name":      tenant.Name,
			"plan":      tenant.Plan,
			"createdAt": tenant.CreatedAt,
		},
		// TODO: Implementar geração de token JWT
		"token": "dummy-token",
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (s *Server) metrics(w http.ResponseWriter, r *http.Request) {
	// TODO: Implementar métricas do Prometheus
	json.NewEncoder(w).Encode(map[string]string{"metrics": "todo"})
}

func (s *Server) handleClassifyEmail(w http.ResponseWriter, r *http.Request) {
	var email entities.Email
	if err := json.NewDecoder(r.Body).Decode(&email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Classificar email
	result, err := s.emailClassifier.ClassifyEmail(r.Context(), &email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Atualizar email com resultados da classificação
	email.Priority = result.Priority
	email.Category = result.Category
	email.Labels = result.Labels
	email.Tasks = result.SuggestedTasks
	email.ProcessedAt = time.Now()

	// Salvar email no banco
	if err := s.emailRepo.Create(r.Context(), &email); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleListEmails(w http.ResponseWriter, r *http.Request) {
	// Extrair filtros da query string
	filters := make(map[string]interface{})
	if category := r.URL.Query().Get("category"); category != "" {
		filters["category"] = category
	}
	if priority := r.URL.Query().Get("priority"); priority != "" {
		filters["priority"] = priority
	}
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		filters["start_date"] = startDate
	}
	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		filters["end_date"] = endDate
	}

	// TODO: Extrair tenant_id do token de autenticação
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		http.Error(w, "tenant_id não fornecido", http.StatusBadRequest)
		return
	}

	emails, err := s.emailRepo.ListByTenant(r.Context(), tenantID, filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(emails)
}

func (s *Server) handleGetEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	email, err := s.emailRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if email == nil {
		http.Error(w, "email não encontrado", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(email)
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	// Extrair parâmetros da query string
	emailID := r.URL.Query().Get("email_id")
	userID := r.URL.Query().Get("user_id")

	// Extrair filtros da query string
	filters := make(map[string]interface{})

	// Paginação
	if page := r.URL.Query().Get("page"); page != "" {
		if pageNum, err := strconv.Atoi(page); err == nil && pageNum > 0 {
			filters["page"] = pageNum
		}
	}
	if pageSize := r.URL.Query().Get("page_size"); pageSize != "" {
		if size, err := strconv.Atoi(pageSize); err == nil && size > 0 && size <= 100 {
			filters["page_size"] = size
		}
	}

	// Filtros de tarefa
	if priority := r.URL.Query().Get("priority"); priority != "" {
		filters["priority"] = entities.Priority(priority)
	}
	if status := r.URL.Query().Get("status"); status != "" {
		filters["status"] = status
	}
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		if date, err := time.Parse(time.RFC3339, startDate); err == nil {
			filters["start_date"] = date
		}
	}
	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		if date, err := time.Parse(time.RFC3339, endDate); err == nil {
			filters["end_date"] = date
		}
	}

	var tasks []*entities.Task
	var err error

	if emailID != "" {
		// Listar tarefas por email
		tasks, err = s.taskRepo.ListByEmail(r.Context(), emailID, filters)
	} else if userID != "" {
		// Listar tarefas pendentes do usuário
		tasks, err = s.taskRepo.ListPendingTasks(r.Context(), userID, filters)
	} else {
		http.Error(w, "email_id ou user_id não fornecido", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(tasks)
}

func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var task entities.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task.ID = id
	if err := s.taskRepo.Update(r.Context(), &task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(task)
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := s.taskRepo.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func mustParseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

func main() {
	log.Printf("Iniciando servidor Email-Filter...")

	server, err := NewServer()
	if err != nil {
		log.Fatalf("Erro ao criar servidor: %v", err)
	}
	defer server.db.Close()
	defer server.emailProcessor.Close()

	// Iniciar processamento de emails em background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.emailProcessor.StartProcessing(ctx); err != nil {
			log.Printf("Erro no processamento de emails: %v", err)
		}
	}()

	server.setupRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Servidor rodando na porta %s", port)
	if err := http.ListenAndServe(":"+port, server.router); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
