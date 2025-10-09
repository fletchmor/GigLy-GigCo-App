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
├── ios-app/              # iOS Mobile Application
│   └── GigCo-Mobile/
│       ├── Views/        # SwiftUI views
│       ├── Services/     # API and business logic services
│       └── Models/       # Data models
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
- **Job completion tracking**:
  - `worker_completed_at` - Timestamp when worker marks job complete
  - `consumer_completed_at` - Timestamp when consumer confirms completion
  - Job is fully completed only when both parties confirm
- **Job status enum**: posted, offer_sent, accepted, rejected, worker_assigned, scheduled, in_progress, completed, paid, review_pending, closed, cancelled, no_worker_available, payment_failed
- **Temporal workflow tracking**: temporal_workflow_id, temporal_run_id, workflow_started_at, workflow_completed_at columns in jobs table

### API Endpoints
- All endpoints prefixed with `/api/v1/`
- Health check available at `/health`
- Comprehensive CRUD operations for all major entities
- Input validation and error handling implemented
- **Job Workflow Endpoints**:
  - `POST /api/v1/jobs/{id}/start` - Start a job (changes status to in_progress)
  - `POST /api/v1/jobs/{id}/complete` - Mark job as complete (dual confirmation system)
  - `POST /api/v1/jobs/{id}/accept` - Accept a job as a gig worker
  - `POST /api/v1/jobs/{id}/reject` - Reject a job offer
- **Job Query Endpoints**:
  - `GET /api/v1/jobs` - Get all jobs (with filters)
  - `GET /api/v1/jobs/available` - Get available jobs for workers (status=posted)
  - `GET /api/v1/jobs/user/{user_id}` - Get jobs for specific user (role-aware)
  - All job responses include consumer and gig_worker user summaries with names

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

### iOS Mobile App
- **Technology**: SwiftUI, iOS 15+
- **Architecture**: MVVM pattern with ObservableObjects
- **API Communication**: Direct URLSession calls to backend API
- **Key Features**:
  - User authentication (login/register)
  - Job browsing and filtering
  - Job acceptance and workflow management (start, complete)
  - Dual completion confirmation system
  - Real-time job status updates
  - Profile management
- **Main Views**:
  - `DashboardView` - Home screen with quick stats and actions
  - `JobListView` - Browse jobs with tabs (All/Available/My Jobs)
  - `JobDetailView` - Detailed job view with action buttons
  - `CreateJobView` - Post new jobs (consumers only)
- **Services**:
  - `APIService` - Handles all HTTP requests to backend
  - `AuthService` - Manages user authentication state
  - `JobService` - Manages job data and local state
- **API Base URL**: http://192.168.22.233:8080/api/v1 (configured in APIService)

### GigCo-Specific Guidelines
- When working with database, always check if services are running: `docker compose ps`
- Use `PGPASSWORD=bamboo psql -h localhost -p 5433 -U postgres -d gigco` for direct DB access
- For API testing, use the Postman collection in `test/` directory
- When adding new endpoints, follow the existing pattern in `api/` directory
- All new database tables should include uuid, created_at, updated_at columns
- Temporal workflows are preferred for any multi-step job processing

### Test User Credentials
- **Worker Account**: worker1@gigco.dev / password123
- **Consumer Account**: testconsumer@gigco.dev / password123
- **Consumer Account**: consumer1@gigco.dev (Alice Johnson) / test123

### Job Workflow States
1. **posted** - Job created by consumer, available for workers
2. **accepted** - Worker has accepted the job
3. **in_progress** - Job has been started by worker
4. **completed** - Both parties have confirmed completion
   - Workers can complete from "accepted" or "in_progress" status
   - Consumers can confirm from "in_progress" or "completed" status
   - Auto-starts job if worker completes from "accepted" status

### Recent Changes (2025-10-09)
- ✅ Fixed iOS navigation - Jobs are now tappable to view details
- ✅ Added dual completion system - Both worker and consumer must confirm
- ✅ Added username display - Shows who posted each job
- ✅ Updated job models to include consumer_name and worker_name fields
- ✅ Fixed authentication to handle NULL password fields
- ✅ Expanded job_status enum to include all workflow states
- ✅ Added temporal tracking columns to jobs table