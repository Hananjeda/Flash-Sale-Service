package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Hananjeda/Flash-Sale-Service/internal/database"
	"github.com/Hananjeda/Flash-Sale-Service/internal/handlers"
	"github.com/Hananjeda/Flash-Sale-Service/internal/redis"
	"github.com/Hananjeda/Flash-Sale-Service/internal/scheduler"
	"github.com/Hananjeda/Flash-Sale-Service/internal/middleware"
)

// Config holds application configuration
type Config struct {
	Port     int
	Database database.Config
	Redis    redis.Config
}

// getEnv returns environment variable value or default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt returns environment variable as integer or default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	return Config{
		Port: getEnvInt("PORT", 8080),
		Database: database.Config{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "flashsale"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: redis.Config{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
	}
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapper, r)
		
		duration := time.Since(start)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapper.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func main() {
	log.Println("Starting Flash Sale Service...")

	// Load configuration
	config := loadConfig()
	log.Printf("Configuration loaded: Port=%d, DB=%s:%d, Redis=%s", 
		config.Port, config.Database.Host, config.Database.Port, config.Redis.Addr)

	// Initialize database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := redis.ConnectRedis()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize scheduler
	scheduler.StartScheduler(db)

	// Create rate limiter
	rateLimiter := middleware.NewRateLimiter(100, 200) // 100 requests per second, burst of 200

	// Setup HTTP routes
	mux := http.NewServeMux()
	
	// API routes
	mux.HandleFunc("/checkout", handlers.CheckoutHandler(db, redisClient))
	mux.HandleFunc("/purchase", handlers.PurchaseHandler(db, redisClient))
	mux.HandleFunc("/health", handlers.HealthCheck(db, redisClient))
	mux.HandleFunc("/stats", handlers.GetStats(db))
	
	// Root route
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "message": "Flash Sale Service is running", "version": "1.0.0"}`))
	})

	// Apply middleware
	finalHandler := corsMiddleware(loggingMiddleware(middleware.RateLimitMiddleware(mux, rateLimiter)))

	// Create HTTP server
	server := &http.Server{
		Addr:         "0.0.0.0:" + strconv.Itoa(config.Port),
		Handler:      finalHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %d", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
