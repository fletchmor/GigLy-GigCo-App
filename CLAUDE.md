# Claude Code Configuration

This file contains configuration and context for Claude Code to help with development tasks.

## Project Information

- **Framework**: Chi v5 (Go HTTP router)
- **Language**: Go 1.23.4
- **Package Manager**: Go Modules
- **Database**: PostgreSQL 17
- **Workflow Engine**: Temporal v1.35.0
- **Containerization**: Docker & Docker Compose

## Common Commands

### Development
- `docker compose up --build` - Start all services (recommended)
- `go run ./cmd/main.go` - Run main API server locally
- `go run ./cmd/worker/main.go` - Run Temporal worker locally
- `go mod tidy` - Update dependencies

### Database Management
- `docker compose exec postgres psql -U postgres -d gigco` - Connect to database
- `PGPASSWORD=bamboo psql -h localhost -p 5433 -U postgres -d gigco` - Connect directly
- `./scripts/reset_and_seed.sh` - Reset and reseed database

### Code Quality
- `go fmt ./...` - Format code
- `go vet ./...` - Static analysis
- `go test ./...` - Run tests

## Project Structure

```
app/
├── cmd/                    # Application entry points
│   ├── main.go            # Main API server
│   └── worker/main.go     # Temporal worker
├── api/                   # HTTP handlers and API logic
│   ├── api.go            # Core API endpoints
│   ├── auth.go           # Authentication endpoints
│   └── job_workflow_handlers.go # Workflow-related endpoints
├── handler/               # Route definitions
├── config/                # Configuration (database, etc.)
├── internal/
│   ├── model/            # Data models and structs
│   ├── middleware/       # HTTP middleware
│   └── temporal/         # Temporal workflows and activities
│       ├── activities/   # Temporal activities
│       ├── workflows/    # Temporal workflows
│       └── client.go     # Temporal client setup
├── scripts/              # Database scripts
│   ├── init.sql         # Complete database schema
│   └── *.sql            # Additional migration scripts
├── templates/            # HTML email templates
├── test/                 # Postman API collections
└── docker-compose.yml    # Development environment
```

## Notes

### Database Schema
- **Main tables**: people, jobs, gigworkers, transactions, schedules
- **Role system**: consumer, gig_worker, admin (stored as enums)
- **UUID support**: Most tables have UUID fields for external references
- **Temporal columns**: created_at/updated_at with automatic triggers

### API Endpoints
- All endpoints prefixed with `/api/v1/`
- Health check available at `/health`
- Comprehensive CRUD operations for all major entities
- Input validation and error handling implemented

### Temporal Workflows
- Job acceptance triggers workflows
- Automatic state management
- UI available at http://localhost:8233
- Workers run in separate containers

### Development Tips
- Use Adminer at http://localhost:8082 for database management
- Import Postman collection from `test/` directory
- Check `docker compose logs` for debugging
- Database runs on port 5433 (host) / 5432 (container)

### Environment Variables
- Database: DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD
- Temporal: TEMPORAL_HOST
- Server: PORT
- All configured in docker-compose.yml

### GigCo-Specific Guidelines
- When working with database, always check if services are running: `docker compose ps`
- Use `PGPASSWORD=bamboo psql -h localhost -p 5433 -U postgres -d gigco` for direct DB access
- For API testing, use the Postman collection in `test/` directory
- When adding new endpoints, follow the existing pattern in `api/` directory
- All new database tables should include uuid, created_at, updated_at columns
- Temporal workflows are preferred for any multi-step job processing