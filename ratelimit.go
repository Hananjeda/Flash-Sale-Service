package middleware

import (
    "net/http"
    "sync"
    "time"
)

type RateLimiter struct {
    rate  int
    burst int
    mutex sync.Mutex
    tokens map[string]int
    lastRefill map[string]time.Time
}

func NewRateLimiter(rate, burst int) *RateLimiter {
    return &RateLimiter{
        rate:  rate,
        burst: burst,
        tokens: make(map[string]int),
        lastRefill: make(map[string]time.Time),
    }
}

func (rl *RateLimiter) Allow(key string) bool {
    rl.mutex.Lock()
    defer rl.mutex.Unlock()

    now := time.Now()
    lastRefill, exists := rl.lastRefill[key]
    if !exists {
        rl.tokens[key] = rl.burst
        rl.lastRefill[key] = now
        return true
    }

    tokensToAdd := int(now.Sub(lastRefill).Seconds()) * rl.rate
    if tokensToAdd > 0 {
        rl.tokens[key] = min(rl.tokens[key]+tokensToAdd, rl.burst)
        rl.lastRefill[key] = now
    }

    if rl.tokens[key] > 0 {
        rl.tokens[key]--
        return true
    }
    return false
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func RateLimitMiddleware(next http.Handler, limiter *RateLimiter) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow(r.RemoteAddr) {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
