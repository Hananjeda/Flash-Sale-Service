# Flash Sale Service Architecture

## System Overview

A high-throughput flash sale service that sells exactly 10,000 items every hour with the following constraints:
- Built with Go (no web frameworks)
- Uses Redis for caching and atomic operations
- Uses PostgreSQL for persistent storage
- Minimal external dependencies
- Handles concurrent requests safely
- Limits users to max 10 items per sale

## Architecture Components

### 1. HTTP Server (Pure Go)
- Uses Go's built-in `net/http` package
- No external web frameworks
- Handles two main endpoints:
  - `POST /checkout?user_id=%user_id%&id=%item_id%`
  - `POST /purchase?code=%code%`

### 2. Database Layer
**PostgreSQL Schema:**
- `sales` - tracks each hourly sale
- `items` - stores item details for each sale
- `checkouts` - persists all checkout attempts
- `purchases` - records successful purchases
- `users` - basic user information

**Redis Data Structures:**
- `sale:{sale_id}:inventory` - atomic counter for remaining items
- `sale:{sale_id}:user:{user_id}:count` - user purchase count per sale
- `checkout:{code}` - temporary checkout session data
- `sale:{sale_id}:active` - sale status flag

### 3. Core Services

**Sale Scheduler:**
- Runs every hour to create new sales
- Generates 10,000 unique items with names and images
- Initializes Redis counters atomically

**Checkout Service:**
- Validates user and item
- Generates unique checkout code
- Stores checkout attempt in both DB and Redis
- Returns checkout code to user

**Purchase Service:**
- Validates checkout code
- Atomically decrements inventory
- Enforces user purchase limits
- Records successful purchase

### 4. Concurrency Strategy

**Atomic Operations:**
- Use Redis DECR for inventory management
- Use Redis INCR for user purchase counting
- Use PostgreSQL transactions for data consistency

**Race Condition Prevention:**
- Redis atomic operations for inventory
- Lua scripts for complex atomic operations
- Database constraints for data integrity

## Data Flow

### Checkout Flow
1. Receive POST /checkout request
2. Validate user_id and item_id
3. Check if sale is active
4. Generate unique checkout code
5. Store checkout in DB and Redis
6. Return checkout code

### Purchase Flow
1. Receive POST /purchase request
2. Validate checkout code in Redis
3. Check user purchase limit (Redis)
4. Atomically decrement inventory (Redis)
5. Record purchase in DB
6. Clean up checkout session
7. Return success/failure

### Sale Initialization Flow
1. Scheduler triggers new sale every hour
2. Generate 10,000 unique items
3. Store items in PostgreSQL
4. Initialize Redis counters
5. Mark sale as active

## Minimal Dependencies

**Required Go Packages:**
- `database/sql` - PostgreSQL driver interface
- `github.com/lib/pq` - PostgreSQL driver (minimal, pure Go)
- `github.com/go-redis/redis/v8` - Redis client (minimal, well-maintained)
- `crypto/rand` - for generating secure codes
- `encoding/json` - for JSON handling
- `net/http` - built-in HTTP server
- `time` - for scheduling

**No Additional Frameworks:**
- No Gin, Echo, or other web frameworks
- No ORM libraries
- No additional middleware packages

## Performance Optimizations

1. **Connection Pooling:** Configure DB and Redis connection pools
2. **Prepared Statements:** Use prepared SQL statements
3. **Redis Pipeline:** Batch Redis operations where possible
4. **Goroutine Pool:** Limit concurrent goroutines to prevent resource exhaustion
5. **Memory Management:** Efficient struct design and minimal allocations

## Error Handling

1. **Database Errors:** Retry with exponential backoff
2. **Redis Errors:** Fallback mechanisms and circuit breaker
3. **Inventory Exhaustion:** Graceful handling when items sold out
4. **Invalid Requests:** Proper HTTP status codes and error messages
5. **System Overload:** Rate limiting and graceful degradation

## Monitoring & Observability

1. **Metrics:** Track throughput, error rates, response times
2. **Logging:** Structured logging for debugging
3. **Health Checks:** Endpoint for service health monitoring
4. **Resource Usage:** Monitor memory, CPU, and connection usage

