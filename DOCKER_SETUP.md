# Docker Setup Instructions

## Prerequisites

### Install Docker
1. **macOS**: Download Docker Desktop from https://docker.com/products/docker-desktop
2. **Linux**: Install Docker Engine following https://docs.docker.com/engine/install/
3. **Windows**: Download Docker Desktop from https://docker.com/products/docker-desktop

### Verify Installation
```bash
docker --version
docker compose --version
```

## Running the Application

### 1. Start the Environment
```bash
# Build and start all services
docker compose up --build

# Or run in background
docker compose up --build -d
```

### 2. Verify Services
```bash
# Check running containers
docker compose ps

# View logs
docker compose logs app
docker compose logs postgres
```

### 3. Test Health Check
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "database": "connected",
  "timestamp": "2025-07-26T17:28:15Z"
}
```

### 4. Test Existing Endpoints
```bash
# Get customer by ID
curl http://localhost:8080/api/v1/customers/1

# Create new user
curl -X POST http://localhost:8080/api/v1/users/create \
  -H "Content-Type: application/json" \
  -d '{"name": "Test User", "address": "123 Test St"}'
```

## Development Workflow

### Making Code Changes
The application container will need to be rebuilt after code changes:
```bash
docker compose down
docker compose up --build
```

### Database Management
```bash
# Connect to PostgreSQL
docker compose exec postgres psql -U postgres -d gigco

# View tables
\dt

# Query sample data
SELECT * FROM customers;
```

### Cleanup
```bash
# Stop services
docker compose down

# Remove volumes (destroys data)
docker compose down -v

# Remove images
docker compose down --rmi all
```

## Troubleshooting

### Common Issues
1. **Port 8080 already in use**: Change the port in docker-compose.yml
2. **Database connection refused**: Ensure PostgreSQL container is healthy
3. **Build failures**: Check Dockerfile and ensure all files are present

### Checking Container Status
```bash
# View container health
docker compose ps

# Check application logs
docker compose logs -f app

# Check database health
docker compose exec postgres pg_isready -U postgres
```

## File Structure
```
├── Dockerfile              # Go application container
├── docker-compose.yml      # Service orchestration
├── .env                    # Environment variables
├── scripts/
│   └── init.sql           # Database initialization
└── templates/             # Email templates (mounted as volume)
```