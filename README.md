# GigCo - Gig Economy Platform

A gig-economy platform where consumers post jobs and gig workers accept them. Built with Go, PostgreSQL, and Docker for local development, with plans for AWS serverless deployment.

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- Git

### Run the Application
```bash
# Clone and navigate to the project
git clone <repository-url>
cd app

# Start the application with Docker
docker compose up --build

# Application will be available at http://localhost:8080
```

## 📋 Current Features

### API Endpoints
- **Health Check**: `GET /health` - Application and database status
- **Customer Management**: `GET /api/v1/customers/{id}` - Retrieve customer by ID  
- **User Creation**: `POST /api/v1/users/create` - Create new users
- **Email Forms**: Basic email form handling (legacy)

### Infrastructure
- **Dockerized Development**: Complete Docker Compose setup
- **PostgreSQL Database**: Version 17 with health checks and data persistence
- **Database Seeding**: Automatic initialization with sample data
- **Health Monitoring**: Built-in health check endpoint

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
- **Database**: PostgreSQL 17
- **Environment**: Docker & Docker Compose
- **Testing**: Postman collections with automated tests

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
DB_HOST=localhost          # postgres for Docker
DB_PORT=5432
DB_NAME=gigco
DB_USER=postgres
DB_PASSWORD=password
DB_SSLMODE=disable

# Server Configuration  
PORT=8080
ENV=development
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

### Schema
- **customers**: User data (id, name, address, timestamps)
- Sample data automatically seeded on startup

### Database Management
```bash
# Connect to database
docker compose exec postgres psql -U postgres -d gigco

# View tables
\dt

# Query sample data
SELECT * FROM customers;
```

## 🗺️ Roadmap

### Phase 1: Foundation (Current - ✅ Complete)
- [x] Docker development environment
- [x] Basic API endpoints  
- [x] Database setup and seeding
- [x] Health monitoring
- [x] Postman test suite

### Phase 2: Core Business Logic (Next)
- [ ] Enhanced database schema (jobs, transactions, schedules)
- [ ] User role management (Consumer, Gig Worker, Admin)
- [ ] Job posting and acceptance workflow
- [ ] Basic transaction system

### Phase 3: Payment Integration
- [ ] Payment provider adapter pattern
- [ ] Clover payment integration
- [ ] Transaction processing and settlement

### Phase 4: Advanced Features
- [ ] Worker scheduling system
- [ ] Notification system
- [ ] Mobile app preparation

### Phase 5: AWS Migration
- [ ] Serverless architecture (Lambda, DynamoDB)
- [ ] API Gateway integration
- [ ] EventBridge and Step Functions

## 📝 Documentation

- `CLAUDE.md` - Development guidance for AI assistants
- `DOCKER_SETUP.md` - Detailed Docker setup instructions
- `implementation-plan.md` - Complete development roadmap
- `requirements.md` - Original project requirements
- `progress-log.md` - Development progress and decisions
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

Currently: **Monolithic Go Application**
- Single binary with embedded HTTP server
- Direct PostgreSQL connection
- Docker Compose for local development

Future: **AWS Serverless**
- Lambda functions with API Gateway
- DynamoDB for data persistence  
- EventBridge for event processing
- Step Functions for workflows

---

**Status**: Phase 1 Complete - Ready for core business logic development  
**Last Updated**: July 26, 2025