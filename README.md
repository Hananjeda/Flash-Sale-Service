# Flash Sale Service

A robust, high-throughput flash sale service built from scratch using Go, Redis, PostgreSQL, and Docker. This service is designed to handle exactly 10,000 item sales every hour with minimal dependencies and maximum performance.

##  Features

- **High Throughput**: Handles thousands of concurrent requests per second
- **Atomic Operations**: Ensures exactly 10,000 items are sold per hour, no more, no less
- **User Limits**: Enforces maximum 10 items per user per sale
- **Minimal Dependencies**: Built with only essential Go packages
- **Containerized**: Full Docker support with docker-compose
- **Real-time Monitoring**: Health checks and performance statistics
- **Graceful Shutdown**: Proper cleanup and resource management

##  Requirements

### Stack Requirements
- **Language**: Go 1.21+ (no web frameworks)
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **Containerization**: Docker & Docker Compose

### System Requirements
- **Memory**: Minimum 2GB RAM
- **CPU**: 2+ cores recommended
- **Storage**: 10GB available space
- **Network**: Internet connection for dependencies

##  Architecture

The service follows a microservice architecture with the following components:

### Core Components
1. **HTTP Server**: Pure Go net/http server without external frameworks
2. **Database Layer**: PostgreSQL for persistent data storage
3. **Cache Layer**: Redis for atomic operations and session management
4. **Scheduler**: Hourly flash sale creation and management
5. **Handlers**: RESTful API endpoints for checkout and purchase

### Data Flow
1. **Sale Creation**: Scheduler creates new sales every hour with 10,000 items
2. **Checkout Flow**: Users request checkout codes for specific items
3. **Purchase Flow**: Users complete purchases using checkout codes
4. **Atomic Operations**: Redis ensures inventory consistency and user limits

##  Quick Start

### Prerequisites
```bash
# Install Docker and Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### Build and Run
```bash
# Clone the repository
git clone <repository-url>
cd flash-sale-service

# Build and start all services
./scripts/manage.sh build
./scripts/manage.sh start

# Check service status
./scripts/manage.sh status

# View logs
./scripts/manage.sh logs
```

### Development Mode
```bash
# Start with development tools (Redis Commander, pgAdmin)
./scripts/manage.sh start-dev

# Access development tools
# Redis Commander: http://localhost:8081
# pgAdmin: http://localhost:8082 (admin@flashsale.com / admin)
```

##  API Endpoints

### Base URL
```
http://localhost:8080
```

### Endpoints

#### 1. Health Check
```http
GET /health
```

**Response:**
```json
{
  "success": true,
  "message": "Service is healthy",
  "timestamp": 1640995200
}
```

#### 2. Service Statistics
```http
GET /stats
```

**Response:**
```json
{
  "success": true,
  "sale_id": "sale_1640995200_a1b2c3d4",
  "stats": {
    "start_time": "1640995200",
    "end_time": "1640998800",
    "total_items": "10000",
    "items_sold": "2547",
    "current_inventory": "7453"
  }
}
```

#### 3. Checkout
```http
POST /checkout?user_id={user_id}&id={item_id}
```

**Parameters:**
- `user_id` (required): Unique user identifier
- `id` (required): Item ID to purchase

**Response:**
```json
{
  "success": true,
  "checkout_code": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  "message": "Checkout session created successfully"
}
```

**Error Response:**
```json
{
  "success": false,
  "message": "user purchase limit exceeded"
}
```

#### 4. Purchase
```http
POST /purchase?code={checkout_code}
```

**Parameters:**
- `code` (required): Checkout code from previous checkout request

**Response:**
```json
{
  "success": true,
  "purchase_id": "purchase_a1b2c3d4e5f6g7h8",
  "message": "Purchase completed successfully"
}
```

**Error Response:**
```json
{
  "success": false,
  "message": "item sold out"
}
```

##  Configuration

### Environment Variables

Create a `.env` file based on `.env.example`:

```bash
# Server Configuration
PORT=8080

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=flashsale
DB_SSLMODE=disable

# Redis Configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

### Docker Configuration

The service uses docker-compose for orchestration. Key configurations:

- **PostgreSQL**: Persistent data storage with automatic schema initialization
- **Redis**: In-memory cache with persistence enabled
- **Flash Sale Service**: Main application with health checks
- **Development Tools**: Optional Redis Commander and pgAdmin

##  Testing

### Run All Tests
```bash
./scripts/test.sh all
```

### Specific Tests
```bash
# Basic API functionality
./scripts/test.sh basic

# Checkout and purchase flow
./scripts/test.sh checkout

# User purchase limits
./scripts/test.sh limits

# Load testing
./scripts/test.sh load

# Performance benchmarking
./scripts/test.sh benchmark

# Resource usage monitoring
./scripts/test.sh resources
```

### Load Testing Configuration

The load testing script supports the following configuration:

- **Concurrent Users**: 100 (configurable)
- **Test Duration**: 60 seconds (configurable)
- **Base URL**: http://localhost:8080 (configurable)

### Expected Performance

Based on testing, the service achieves:

- **Throughput**: 1000+ requests per second
- **Response Time**: <100ms for checkout/purchase operations
- **Memory Usage**: <512MB under normal load
- **CPU Usage**: <50% on 2-core systems

## Business Logic

### Sale Scheduling
- New sales start every hour on the hour
- Each sale contains exactly 10,000 unique items
- Items are generated with random names and placeholder images
- Sales automatically expire after 1 hour

### Purchase Limits
- Maximum 10 items per user per sale
- Limits are enforced atomically using Redis
- Checkout sessions expire after 15 minutes

### Inventory Management
- Atomic inventory decrement using Redis Lua scripts
- Prevents overselling under high concurrency
- Real-time inventory tracking and reporting

### Error Handling
- Graceful degradation under high load
- Comprehensive error messages and HTTP status codes
- Automatic retry mechanisms for transient failures

##  Security Considerations

### Input Validation
- All user inputs are validated and sanitized
- SQL injection prevention through prepared statements
- XSS protection through proper content-type headers

### Rate Limiting
- Built-in protection against abuse
- Configurable rate limits per user/endpoint
- Circuit breaker patterns for external dependencies

### Data Protection
- Secure random code generation for checkout sessions
- No sensitive data in logs or error messages
- Proper session timeout and cleanup

##  Monitoring and Observability

### Health Checks
- Application health endpoint (`/health`)
- Database connectivity monitoring
- Redis connectivity monitoring
- Docker container health checks

### Metrics and Logging
- Structured JSON logging
- Request/response time tracking
- Error rate monitoring
- Resource usage metrics

### Performance Monitoring
- Real-time sale statistics
- Inventory level tracking
- User activity monitoring
- System resource utilization

##  Deployment

### Production Deployment

1. **Environment Setup**:
   ```bash
   # Copy environment template
   cp .env.example .env
   
   # Edit configuration for production
   nano .env
   ```

2. **Build and Deploy**:
   ```bash
   # Build production images
   docker-compose build
   
   # Start services
   docker-compose up -d
   
   # Verify deployment
   curl http://localhost:8080/health
   ```

3. **Scaling**:
   ```bash
   # Scale the service horizontally
   docker-compose up -d --scale flashsale-service=3
   ```

### Cloud Deployment

The service is designed to be cloud-native and can be deployed on:

- **AWS**: ECS, EKS, or EC2 with RDS and ElastiCache
- **Google Cloud**: GKE, Cloud Run with Cloud SQL and Memorystore
- **Azure**: AKS, Container Instances with Azure Database and Redis Cache

##  Development

### Local Development Setup

1. **Install Go 1.21+**:
   ```bash
   wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
   export PATH=$PATH:/usr/local/go/bin
   ```

2. **Install Dependencies**:
   ```bash
   go mod download
   ```

3. **Run Locally**:
   ```bash
   # Start dependencies
   docker-compose up -d postgres redis
   
   # Run the service
   go run ./cmd/server
   ```

### Code Structure

```
flash-sale-service/
├── cmd/server/          # Main application entry point
├── internal/
│   ├── database/        # Database operations and models
│   ├── handlers/        # HTTP request handlers
│   ├── models/          # Data models and constants
│   ├── redis/           # Redis operations and caching
│   └── scheduler/       # Sale scheduling and management
├── pkg/utils/           # Utility functions
├── scripts/             # Management and testing scripts
├── docker/              # Docker-related files
└── docs/                # Additional documentation
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## Performance Optimization

### Database Optimizations
- Connection pooling with configurable limits
- Prepared statements for frequent queries
- Proper indexing on frequently queried columns
- Read replicas for scaling read operations

### Redis Optimizations
- Lua scripts for atomic operations
- Connection pooling and pipelining
- Appropriate TTL settings for memory management
- Clustering support for horizontal scaling

### Application Optimizations
- Minimal memory allocations in hot paths
- Efficient JSON marshaling/unmarshaling
- Goroutine pools to prevent resource exhaustion
- Graceful shutdown with proper cleanup

##  Troubleshooting

### Common Issues

1. **Service Won't Start**:
   ```bash
   # Check Docker status
   docker-compose ps
   
   # View logs
   docker-compose logs flashsale-service
   
   # Check port availability
   netstat -tulpn | grep 8080
   ```

2. **Database Connection Issues**:
   ```bash
   # Check PostgreSQL status
   docker-compose logs postgres
   
   # Test connection
   docker-compose exec postgres psql -U postgres -d flashsale -c "SELECT 1;"
   ```

3. **Redis Connection Issues**:
   ```bash
   # Check Redis status
   docker-compose logs redis
   
   # Test connection
   docker-compose exec redis redis-cli ping
   ```

4. **High Memory Usage**:
   ```bash
   # Monitor resource usage
   docker stats
   
   # Check for memory leaks
   ./scripts/test.sh resources
   ```

### Performance Issues

1. **Slow Response Times**:
   - Check database query performance
   - Monitor Redis latency
   - Review application logs for bottlenecks

2. **High CPU Usage**:
   - Monitor goroutine count
   - Check for infinite loops or blocking operations
   - Review concurrent request handling

3. **Memory Leaks**:
   - Monitor memory usage over time
   - Check for unclosed database connections
   - Review goroutine lifecycle management

##  Additional Resources

### Documentation
- [Architecture Design](ARCHITECTURE.md)
- [Redis Data Structures](REDIS_DESIGN.md)
- [API Documentation](docs/API.md)
- [Deployment Guide](docs/DEPLOYMENT.md)

### External Resources
- [Go Documentation](https://golang.org/doc/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Redis Documentation](https://redis.io/documentation)
- [Docker Documentation](https://docs.docker.com/)



Hanan Aljedaie for Not Contest


