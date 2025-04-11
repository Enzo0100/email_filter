package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/enzo010/email-filter/internal/application/services"
	"github.com/enzo010/email-filter/internal/domain/entities"
	"github.com/enzo010/email-filter/internal/infrastructure/middleware"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	emailClassifier *services.EmailClassifier
	router         *mux.Router
}

func NewServer() (*Server, error) {
	// Carregar variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Printf("Arquivo .env não encontrado: %v", err)
	}

	// Inicializar classificador
	emailClassifier := services.NewEmailClassifier()

	// Inicializar router
	router := mux.NewRouter()

	return &Server{
		emailClassifier: emailClassifier,
		router:         router,
	}, nil
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "http://localhost:3000" // Fallback para desenvolvimento
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-Tenant-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "300")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) setupRoutes() {
	// Configurar rate limiter (10 requisições por segundo por IP)
	rateLimiter := middleware.NewRateLimiter(10, 10)

	// Aplicar middlewares globais
	s.router.Use(corsMiddleware)
	s.router.Use(middleware.MetricsMiddleware)
	s.router.Use(middleware.RateLimitMiddleware(rateLimiter))

	// Endpoints públicos
	s.router.HandleFunc("/health", s.healthCheck).Methods("GET")
	s.router.Handle("/metrics", promhttp.Handler())

	// API v1
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Endpoint de classificação
	api.HandleFunc("/classify", s.handleClassifyEmail).Methods("POST")
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (s *Server) handleClassifyEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var email entities.Email
	if err := json.NewDecoder(r.Body).Decode(&email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Mock classification result for now
	result := &entities.ClassificationResult{
		Priority: "high",
		Category: "meeting",
		Labels: []string{"urgent", "report", "meeting"},
		SuggestedTasks: []*entities.Task{
			{
				Title: "Review quarterly report",
				Description: "Review the quarterly report before tomorrow's meeting",
				Priority: "high",
			},
			{
				Title: "Prepare meeting notes",
				Description: "Prepare notes for tomorrow's urgent meeting",
				Priority: "medium",
			},
		},
	}

	json.NewEncoder(w).Encode(result)
}

func main() {
	log.Printf("Iniciando serviço de classificação de emails...")

	server, err := NewServer()
	if err != nil {
		log.Fatalf("Erro ao criar servidor: %v", err)
	}

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
