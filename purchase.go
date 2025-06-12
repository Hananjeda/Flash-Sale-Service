package handlers

import (
    "encoding/json"
    "net/http"
    "github.com/Hananjeda/Flash-Sale-Service/internal/database"
    "github.com/Hananjeda/Flash-Sale-Service/internal/redis"
)

func PurchaseHandler(db *sql.DB, redisClient *redis.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        checkoutCode := r.URL.Query().Get("code")

        if checkoutCode == "" {
            http.Error(w, "Missing checkout code", http.StatusBadRequest)
            return
        }

        // Retrieve checkout session from Redis
        userID, itemID, err := redis.GetCheckoutSession(redisClient, checkoutCode)
        if err != nil {
            http.Error(w, "Invalid or expired checkout code", http.StatusBadRequest)
            return
        }

        // Perform atomic inventory decrement
        decremented, err := redis.DecrementInventory(redisClient, itemID)
        if err != nil {
            http.Error(w, "Error processing purchase", http.StatusInternalServerError)
            return
        }

        if !decremented {
            http.Error(w, "Item sold out", http.StatusConflict)
            return
        }

        // Record the purchase in the database
        purchaseID, err := recordPurchase(db, userID, itemID)
        if err != nil {
            http.Error(w, "Error recording purchase", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "success":     true,
            "purchase_id": purchaseID,
            "message":     "Purchase completed successfully",
        })
    }
}

func recordPurchase(db *sql.DB, userID, itemID string) (string, error) {
    // Implementation for recording the purchase in the database
    // This is a placeholder and should be implemented with actual logic
    return "purchase_a1b2c3d4e5f6g7h8", nil
}
