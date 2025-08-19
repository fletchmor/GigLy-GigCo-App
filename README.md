# GigCo - Gig Economy Platform

A comprehensive gig-economy platform that connects consumers with gig workers for various services. Built with Go, PostgreSQL, Temporal workflow engine, and Docker for local development.

## 🌟 Key Features

- **User Management**: Consumer, gig worker, and admin role-based system
- **Job Management**: Complete job posting, acceptance, and completion workflow
- **Workflow Automation**: Temporal-powered job processing and state management
- **Payment Processing**: Transaction handling with settlement batching
- **Scheduling System**: Worker availability and job scheduling
- **Notification System**: Real-time notifications for job updates
- **Review System**: Job ratings and feedback

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- Git
- Go 1.23+ (for local development)
- PostgreSQL client (optional, for direct database access)

### Run the Application
```bash
# Clone and navigate to the project
git clone <repository-url>
cd app

# Start all services with Docker
docker compose up --build

# Services will be available at:
# - Main API: http://localhost:8080
# - Temporal UI: http://localhost:8233
# - Database Admin: http://localhost:8082
# - PostgreSQL: localhost:5433
```

## 📋 Current Features

### API Endpoints

#### Core System
- **Health Check**: `GET /health` - Application and database status
- **User Registration**: `POST /api/v1/auth/register` - Register users with role selection
- **Customer Management**: `GET /api/v1/customers/{id}` - Retrieve customer by ID (legacy)

#### GigWorker Management
- **List Workers**: `GET /api/v1/gigworkers` - List gig workers with filtering
- **Get Worker**: `GET /api/v1/gigworkers/{id}` - Get specific gig worker details
- **Create Worker**: `POST /api/v1/gigworkers/create` - Register new gig workers

#### Job Management
- **List Jobs**: `GET /api/v1/jobs` - List available jobs with filtering
- **Get Job**: `GET /api/v1/jobs/{id}` - Get specific job details
- **Create Job**: `POST /api/v1/jobs/create` - Post new jobs
- **Accept Job**: `POST /api/v1/jobs/{id}/accept` - Accept jobs (triggers workflow)

#### Financial System
- **Create Transaction**: `POST /api/v1/transactions/create` - Process payments

#### Scheduling
- **Create Schedule**: `POST /api/v1/schedules/create` - Manage worker availability

### Infrastructure
- **Dockerized Development**: Complete Docker Compose setup with 5 services
- **PostgreSQL Database**: Version 17 with comprehensive schema and health checks
- **Temporal Workflows**: Automated job processing and state management
- **Database Administration**: Adminer web interface for database management
- **Workflow Monitoring**: Temporal UI for workflow visualization
- **Health Monitoring**: Built-in health check endpoints

## 🛠️ Development

### Project Structure
```
├── cmd/main.go              # Application entry point
├── api/                     # HTTP handlers and API logic
├── handler/                 # Route definitions  
├── config/                  # Database configuration
├── internal/
│   ├── model/              # Data models
│   └── middleware/         # HTTP middleware
├── scripts/init.sql        # Database initialization
├── templates/              # HTML email templates
├── test/                   # Postman API collections
├── docker-compose.yml      # Development environment
└── Dockerfile             # Application container
```

### Tech Stack
- **Language**: Go 1.23.4
- **Router**: Chi v5
- **Database**: PostgreSQL 17 with comprehensive schema
- **Workflow Engine**: Temporal v1.35.0
- **Environment**: Docker & Docker Compose
- **Testing**: Postman collections with comprehensive API tests
- **Database Admin**: Adminer web interface

### Running Locally

#### With Docker (Recommended)
```bash
# Start all services
docker compose up --build

# Run in background
docker compose up --build -d

# Check service status
docker compose ps

# View logs
docker compose logs app
docker compose logs postgres

# Stop services
docker compose down
```

#### Manual Setup
```bash
# Install dependencies
go mod tidy

# Set up environment variables
cp .env.example .env
# Edit .env with your database settings

# Run the application (requires PostgreSQL running)
go run ./cmd/main.go
```

### Environment Variables
```bash
# Database Configuration
DB_HOST=postgres           # For Docker, localhost for manual setup
DB_PORT=5432               # Container port (5433 for host access)
DB_NAME=gigco
DB_USER=postgres
DB_PASSWORD=bamboo
DB_SSLMODE=disable

# Server Configuration  
PORT=8080

# Temporal Configuration
TEMPORAL_HOST=temporal:7233
```

## 🧪 Testing

### Postman Collection
Comprehensive API testing suite available in `test/` directory:

```bash
# Import into Postman:
test/GigCo-API.postman_collection.json
test/GigCo-Local.postman_environment.json
```

**Test Coverage:**
- Health check validation
- Customer data retrieval
- User creation with validation
- Request chaining and data consistency
- Error scenario handling

### Manual API Testing
```bash
# Health check
curl http://localhost:8080/health

# Get customer
curl http://localhost:8080/api/v1/customers/1

# Create user
curl -X POST http://localhost:8080/api/v1/users/create \
  -H "Content-Type: application/json" \
  -d '{"name": "Test User", "address": "123 Main St"}'
```

## 📊 Database

### Schema Overview
Comprehensive gig-economy database with 15+ tables:

#### Core Tables
- **people**: Users (consumers, gig workers, admins) with roles and verification
- **jobs**: Job postings with status tracking and location data
- **gigworkers**: Enhanced worker profiles with skills and availability
- **transactions**: Payment processing with settlement batching
- **schedules**: Worker availability and job scheduling

#### Supporting Tables
- **notifications**: In-app notification system
- **job_reviews**: Rating and review system
- **payment_providers**: Multi-provider payment support
- **worker_profiles**: Extended worker information
- **worker_templates**: Service category templates
- **worker_services**: Worker-to-service mappings

### Database Access Options

#### Web Interface (Recommended)
```bash
# Access Adminer at http://localhost:8082
# Server: postgres
# Username: postgres
# Password: bamboo
# Database: gigco
```

#### Command Line
```bash
# Connect via Docker
docker compose exec postgres psql -U postgres -d gigco

# Connect directly (requires postgres client)
PGPASSWORD=bamboo psql -h localhost -p 5433 -U postgres -d gigco

# View all tables
\dt

# Sample queries
SELECT * FROM people WHERE role = 'gig_worker';
SELECT * FROM jobs ORDER BY created_at DESC LIMIT 5;
```

## 🗺️ Development Status

### ✅ Completed Features

#### Foundation & Infrastructure
- [x] Docker development environment with 5 services
- [x] Comprehensive PostgreSQL schema (15+ tables)
- [x] Health monitoring endpoints
- [x] Database administration interface

#### Core Business Logic
- [x] User role management (Consumer, Gig Worker, Admin)
- [x] Complete job posting and acceptance workflow
- [x] GigWorker management system
- [x] Transaction tracking system
- [x] Worker scheduling system

#### Workflow Automation
- [x] Temporal workflow integration
- [x] Automated job processing
- [x] Job state management
- [x] Workflow monitoring UI

#### API & Testing
- [x] Comprehensive REST API
- [x] Postman test collection
- [x] Input validation and error handling
- [x] Database seeding and fixtures

### 🚧 In Development
- [ ] Payment provider integration
- [ ] Email notification system
- [ ] Advanced worker matching algorithms
- [ ] Mobile API optimizations

### 📋 Future Enhancements
- [ ] Real-time notifications
- [ ] Advanced reporting dashboard
- [ ] Mobile app support
- [ ] Multi-language support
- [ ] AWS serverless migration

## 📝 Documentation

- **[docs/](./docs/)** - Complete technical documentation index\n- `CLAUDE.md` - Development guidance for AI assistants
- `DOCKER_SETUP.md` - Detailed Docker setup instructions
- **[docs/implementation-plan.md](./docs/implementation-plan.md)** - Complete development roadmap
- **[docs/requirements.md](./docs/requirements.md)** - Original project requirements
- **[docs/progress-log.md](./docs/progress-log.md)** - Development progress and decisions
- `test/README.md` - Postman testing documentation

## 🤝 Contributing

1. Ensure Docker is running
2. Make changes to the code
3. Test with Postman collection
4. Rebuild containers: `docker compose up --build`
5. Verify health check: `curl http://localhost:8080/health`

## 📞 Support

- Check `DOCKER_SETUP.md` for troubleshooting
- Review `progress-log.md` for known issues and solutions
- Use Postman collection for API validation

## 🏗️ Architecture

### Current Architecture: **Microservices with Workflow Engine**

#### Services
- **Main API**: Go HTTP server with Chi router
- **Worker Service**: Temporal workflow worker
- **PostgreSQL**: Primary data store
- **Temporal Server**: Workflow orchestration
- **Temporal UI**: Workflow monitoring interface
- **Adminer**: Database administration

#### Data Flow
1. **API Requests** → Main application server
2. **Job Creation** → Database + Temporal workflow
3. **Job Acceptance** → Workflow state transition
4. **Background Processing** → Temporal workers
5. **Database Updates** → Automatic via workflows

### Service Communication
- **HTTP REST API**: Client ↔ Main application
- **Temporal gRPC**: Main application ↔ Workflow engine
- **PostgreSQL**: All services ↔ Database
- **Docker Network**: Internal service communication

### Key Benefits
- **Reliability**: Temporal ensures workflow completion
- **Scalability**: Independent service scaling
- **Observability**: Workflow visibility and monitoring
- **Fault Tolerance**: Automatic retry and error handling

---

**Status**: Core Platform Complete - Production Ready  
**Last Updated**: August 19, 2025