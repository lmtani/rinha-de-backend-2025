# Payment Processing System

This is a payment processing system built with clean architecture principles, using:

- Go for backend services
- PostgreSQL for persistent storage
- Redis for queue and temporary storage
- Nginx for load balancing

## Architecture

The system follows clean architecture principles with:

- **Domain Layer**: Core business logic and entities
- **Use Case Layer**: Application-specific business rules
- **Interface Adapters**: Repositories, controllers, etc.
- **Framework & Drivers**: External frameworks and tools

## Database Setup

The PostgreSQL database is automatically initialized when the container starts up. The initialization script in `docker/postgres/init/01-init-schema.sql` creates:

- The `payments` table
- Required indexes

## Components

- **Repository**: Uses PostgreSQL for persistent storage of payment records
- **Queue**: Uses Redis lists for payment processing queue
- **Store**: Uses Redis for UUID deduplication with TTL

## Running the Application

```bash
# Start all services
docker compose up -d

# Check logs
docker compose logs -f

# Stop all services
docker compose down
```

## Environment Variables

### Database
- `DATABASE_URL`: PostgreSQL connection string
- `DATABASE_MAX_CONNECTIONS`: Maximum number of DB connections
- `DATABASE_CONNECT_TIMEOUT`: Timeout for DB connection

### Redis
- `REDIS_URL`: Redis connection URL
- `REDIS_POOL_SIZE`: Maximum number of Redis connections
- `REDIS_QUEUE_KEY`: Key for the payment queue
- `REDIS_UUID_TTL`: Time-to-live for UUID cache

### API
- `SERVER_PORT`: Port for the API server
- `SERVER_READ_TIMEOUT`: Timeout for reading requests
- `SERVER_WRITE_TIMEOUT`: Timeout for writing responses

## API Endpoints

- **POST /payments**: Request a payment processing
- **GET /payments-summary**: Get summary of processed payments
  - Optional query params: `from` and `to` in ISO 8601 format (UTC)
- **GET /health**: Health check endpoint
