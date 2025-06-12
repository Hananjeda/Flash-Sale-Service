package handlers

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/Hananjeda/Flash-Sale-Service/internal/database"
    "github.com/Hananjeda/Flash-Sale-Service/internal/redis"
)

func HealthCheck(db *database.DB, redisClient *redis.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        health := struct {
            Status    string `json:"status"`
            Timestamp int64  `json:"timestamp"`
            Database  string `json:"database"`
            Redis     string `json:"redis"`
        }{
            Status:    "OK",
            Timestamp: time.Now().Unix(),
            Database:  "OK",
            Redis:     "OK",
        }

        if err := db.Ping(); err != nil {
            health.Status = "ERROR"
            health.Database = "ERROR"
        }

        if err := redisClient.Ping(); err != nil {
            health.Status = "ERROR"
            health.Redis = "ERROR"
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(health)
    }
}
