package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/enzo010/email-filter/internal/infrastructure/auth"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/time/rate"
)

// Logger para logging estruturado
type Logger struct {
	*log.Logger
}

func (l *Logger) logRequest(r *http.Request, code int, duration time.Duration) {
	log.Printf(`{"timestamp":"%s","method":"%s","path":"%s","status":%d,"duration":"%s","user_agent":"%s"}`,
		time.Now().Format(time.RFC3339),
		r.Method,
		r.URL.Path,
		code,
		duration.String(),
		r.UserAgent(),
	)
}

// RateLimiter implementa um token bucket rate limiter por IP
type RateLimiter struct {
	visitors   map[string]*visitorInfo
	mu         sync.RWMutex
	rate       rate.Limit
	burst      int
	expiration time.Duration
}

type visitorInfo struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	rl := &RateLimiter{
		visitors:   make(map[string]*visitorInfo),
		rate:       r,
		burst:      b,
		expiration: time.Hour,
	}
	go rl.cleanupVisitors() // Inicia a limpeza periódica
	return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	info, exists := rl.visitors[ip]
	if !exists {
		info = &visitorInfo{
			limiter:  rate.NewLimiter(rl.rate, rl.burst),
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = info
	}
	info.lastSeen = time.Now()
	return info.limiter
}

// Limpa visitantes antigos periodicamente
func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(rl.expiration)
		rl.mu.Lock()
		for ip, info := range rl.visitors {
			if time.Since(info.lastSeen) > rl.expiration {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_duration_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"path"})

	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
			ConstLabels: prometheus.Labels{
				"service": "email-filter",
			},
		},
		[]string{"method", "path", "status"},
	)
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	TenantIDKey contextKey = "tenant_id"
	EmailKey    contextKey = "email"
	RoleKey     contextKey = "role"
)

// RateLimitMiddleware limita o número de requisições por IP
func RateLimitMiddleware(limiter *RateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if !limiter.getLimiter(ip).Allow() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Limite de requisições excedido",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// TenantValidationMiddleware verifica se o tenant_id é válido
func TenantValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID, ok := r.Context().Value(TenantIDKey).(int)
		if !ok || tenantID <= 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Tenant ID inválido",
			})
			return
		}

		// Aqui você pode adicionar mais validações específicas do tenant
		// Por exemplo, verificar se o tenant está ativo, tem permissões, etc.

		next.ServeHTTP(w, r)
	})
}

// MetricsMiddleware coleta métricas do Prometheus
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap ResponseWriter para capturar o status code
		wrapped := wrapResponseWriter(w)

		next.ServeHTTP(wrapped, r)

		// Registrar duração
		duration := time.Since(start).Seconds()
		httpDuration.WithLabelValues(r.URL.Path).Observe(duration)

		// Registrar request
		requestsTotal.WithLabelValues(r.Method, r.URL.Path, wrapped.status()).Inc()
	})
}

// CORSMiddleware adiciona os headers CORS necessários
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware verifica e valida o token JWT
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger := &Logger{log.Default()}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			logger.logRequest(r, http.StatusUnauthorized, time.Since(start))
			return
		}

		// Extrair token do header (Bearer token)
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			logger.logRequest(r, http.StatusUnauthorized, time.Since(start))
			return
		}

		claims, err := auth.ValidateToken(tokenParts[1])
		if err != nil {
			statusCode := http.StatusUnauthorized
			var message string

			switch err {
			case auth.ErrExpiredToken:
				message = "Token expirado"
			case auth.ErrInvalidToken:
				message = "Token inválido"
			default:
				message = "Erro na autenticação"
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			json.NewEncoder(w).Encode(map[string]string{
				"error": message,
			})
			logger.logRequest(r, statusCode, time.Since(start))
			return
		}

		// Adicionar claims ao contexto
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, TenantIDKey, claims.TenantID)
		ctx = context.WithValue(ctx, EmailKey, claims.Email)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)

		// Log de sucesso
		logger.logRequest(r, http.StatusOK, time.Since(start))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// responseWriter é um wrapper para http.ResponseWriter que captura o status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) status() string {
	return http.StatusText(rw.statusCode)
}
