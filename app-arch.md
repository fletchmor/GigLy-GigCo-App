# GigCo Application Architecture Documentation

## Overview

GigCo is a gig economy platform that connects consumers with gig workers for various services. The application is built using a microservices architecture with Docker containerization, featuring a Go-based API backend and PostgreSQL database.

## System Architecture

### High-Level Architecture

```mermaid
graph TB
    subgraph "Client Layer"
        API[REST API Clients]
        PM[Postman Tests]
    end
    
    subgraph "Application Layer"
        APP[Go Application<br/>Port 8080]
    end
    
    subgraph "Data Layer"
        PG[(PostgreSQL<br/>Port 5432)]
        ADM[Adminer<br/>Port 8082]
    end
    
    subgraph "Infrastructure"
        DC[Docker Compose]
        NET[gigco-network]
    end
    
    API --> APP
    PM --> APP
    APP --> PG
    ADM --> PG
    
    DC -.manages.-> APP
    DC -.manages.-> PG
    DC -.manages.-> ADM
    
    APP -.connected via.-> NET
    PG -.connected via.-> NET
    ADM -.connected via.-> NET
```

## Container Architecture

The application uses Docker Compose to orchestrate three main services:

1. **App Service**: The main Go application
   - Exposed on port 8080
   - Connected to PostgreSQL database
   - Mounts templates directory for views

2. **PostgreSQL Service**: Database server
   - PostgreSQL 17 Alpine version
   - Exposed on host port 5433 (maps to container port 5432)
   - Persistent volume for data storage
   - Health checks configured

3. **Adminer Service**: Database administration tool
   - Web-based database management
   - Exposed on port 8082
   - Connected to PostgreSQL via Docker network

## Database Schema

### Entity Relationship Diagram

```mermaid
erDiagram
    PEOPLE ||--o{ JOBS : "posts as consumer"
    PEOPLE ||--o{ JOBS : "accepts as worker"
    PEOPLE ||--o{ TRANSACTIONS : "pays as consumer"
    PEOPLE ||--o{ TRANSACTIONS : "receives as worker"
    PEOPLE ||--o{ SCHEDULES : "has availability"
    PEOPLE ||--o{ USER_PAYMENT_METHODS : "has"
    PEOPLE ||--o{ NOTIFICATIONS : "receives"
    PEOPLE ||--o{ NOTIFICATION_PREFERENCES : "configures"
    PEOPLE ||--o{ JOB_REVIEWS : "writes"
    PEOPLE ||--o{ JOB_REVIEWS : "receives"
    
    JOBS ||--o{ TRANSACTIONS : "generates"
    JOBS ||--o{ JOB_REVIEWS : "has"
    JOBS ||--o{ SCHEDULES : "may book"
    JOBS ||--o{ NOTIFICATIONS : "triggers"
    
    PAYMENT_PROVIDERS ||--o{ USER_PAYMENT_METHODS : "provides"
    PAYMENT_PROVIDERS ||--o{ SETTLEMENT_BATCHES : "processes"
    
    SETTLEMENT_BATCHES ||--o{ TRANSACTIONS : "contains"
    
    TRANSACTIONS ||--o{ NOTIFICATIONS : "triggers"

    PEOPLE {
        int id PK
        uuid uuid UK
        string email UK
        string name
        string phone
        text address
        decimal latitude
        decimal longitude
        string place_id
        enum role
        boolean is_active
        boolean email_verified
        boolean phone_verified
        timestamp created_at
        timestamp updated_at
    }

    JOBS {
        int id PK
        uuid uuid UK
        int consumer_id FK
        int gig_worker_id FK
        string title
        text description
        string category
        text location_address
        decimal location_latitude
        decimal location_longitude
        decimal estimated_duration_hours
        decimal pay_rate_per_hour
        decimal total_pay
        enum status
        timestamp scheduled_start
        timestamp scheduled_end
        timestamp actual_start
        timestamp actual_end
        text notes
        timestamp created_at
        timestamp updated_at
    }

    TRANSACTIONS {
        int id PK
        uuid uuid UK
        int job_id FK
        int consumer_id FK
        int gig_worker_id FK
        decimal amount
        string currency
        enum status
        string payment_intent_id
        string payment_method
        timestamp escrow_released_at
        decimal processing_fee
        decimal net_amount
        int settlement_batch_id FK
        text notes
        timestamp created_at
        timestamp updated_at
    }

    SCHEDULES {
        int id PK
        uuid uuid UK
        int gig_worker_id FK
        string title
        timestamp start_time
        timestamp end_time
        boolean is_available
        int job_id FK
        string recurring_pattern
        timestamp recurring_until
        text notes
        timestamp created_at
        timestamp updated_at
    }

    CUSTOMERS {
        int id PK
        string name
        text address
        timestamp created_at
        timestamp updated_at
    }
```

### Core Entity Types

The system uses several PostgreSQL ENUM types for data integrity:

- **user_role**: 'consumer', 'gig_worker', 'admin'
- **job_status**: 'posted', 'accepted', 'in_progress', 'completed', 'cancelled'
- **transaction_status**: 'pending', 'completed', 'failed', 'refunded'
- **notification_type**: 'job_posted', 'job_accepted', 'job_completed', 'payment_received', 'system_message'
- **notification_status**: 'unread', 'read', 'archived'

## API Endpoints

Based on the Postman collection, the application exposes the following REST API endpoints:

### Health & Status
- `GET /health` - Application health check
- `GET /` - Email form (legacy endpoint)

### User Management
- `GET /api/v1/customers/{id}` - Get customer by ID
- `POST /api/v1/users/create` - Create new user
- `POST /api/v1/auth/register` - Register new user with authentication

### Job Management
- `GET /api/v1/jobs` - List all jobs
- `GET /api/v1/jobs/{id}` - Get job by ID
- `POST /api/v1/jobs/create` - Create new job
- `POST /api/v1/jobs/{id}/accept` - Accept a job

### Schedule Management
- `POST /api/v1/schedules/create` - Create schedule entry

### Transaction Management
- `POST /api/v1/transactions/create` - Create transaction

## Application Flow Diagrams

### Job Lifecycle Flow

```mermaid
sequenceDiagram
    participant C as Consumer
    participant API as GigCo API
    participant DB as PostgreSQL
    participant W as Worker
    participant N as Notification Service

    C->>API: POST /api/v1/jobs/create
    API->>DB: Insert job (status: 'posted')
    DB-->>API: Job created
    API-->>C: 201 Created
    API->>N: Trigger job_posted notification
    N->>W: Notify available workers

    W->>API: GET /api/v1/jobs
    API->>DB: Query available jobs
    DB-->>API: Return job list
    API-->>W: 200 OK (jobs array)

    W->>API: POST /api/v1/jobs/{id}/accept
    API->>DB: Update job (gig_worker_id, status: 'accepted')
    DB-->>API: Job updated
    API-->>W: 200 OK
    API->>N: Trigger job_accepted notification
    N->>C: Notify consumer

    Note over W,C: Job execution happens

    W->>API: Update job status to 'completed'
    API->>DB: Update job status
    API->>API: Create transaction record
    API->>N: Trigger job_completed notification
    N->>C: Notify consumer
    N->>W: Notify worker
```

### User Registration Flow

```mermaid
sequenceDiagram
    participant U as User
    participant API as GigCo API
    participant DB as PostgreSQL
    participant NP as Notification Preferences

    U->>API: POST /api/v1/auth/register
    Note over U,API: {name, email, password, role, address}
    
    API->>API: Validate input data
    API->>DB: Check email uniqueness
    DB-->>API: Email available
    
    API->>DB: Insert into people table
    DB-->>API: User created with ID
    
    API->>NP: Create default notification preferences
    NP->>DB: Insert preferences for all notification types
    DB-->>NP: Preferences created
    
    API-->>U: 201 Created
    Note over U: User can now login and use platform
```

### Payment Transaction Flow

```mermaid
sequenceDiagram
    participant C as Consumer
    participant API as GigCo API
    participant DB as PostgreSQL
    participant PP as Payment Provider
    participant W as Worker
    participant SB as Settlement Batch

    Note over C: Job completed
    
    C->>API: POST /api/v1/transactions/create
    Note over C,API: {job_id, consumer_id, worker_id, amount}
    
    API->>DB: Create transaction (status: 'pending')
    DB-->>API: Transaction created
    
    API->>PP: Process payment
    PP-->>API: Payment intent created
    
    API->>DB: Update transaction with payment_intent_id
    
    PP->>API: Webhook: Payment completed
    API->>DB: Update transaction (status: 'completed')
    
    Note over API,SB: Daily settlement process
    
    SB->>DB: Create settlement batch
    SB->>DB: Associate transactions with batch
    SB->>W: Transfer funds to worker
    
    API->>DB: Update transaction (escrow_released_at)
    API->>W: Payment notification
```

## Security & Infrastructure Considerations

### Network Security
- All services communicate through an isolated Docker network (`gigco-network`)
- Database is not directly exposed to the internet (only through mapped port for development)
- Health checks ensure service availability

### Data Persistence
- PostgreSQL data stored in Docker volume (`postgres_data`)
- Database initialization script runs on first startup
- Automatic timestamps and UUID generation for audit trails

### Environment Configuration
- Environment variables used for database configuration
- SSL mode disabled for local development
- Restart policies ensure service availability

## Development & Testing

### API Testing
The Postman collection provides comprehensive test coverage including:
- Health checks
- CRUD operations for all major entities
- Error handling scenarios
- Response time validation
- Data structure validation

### Database Administration
- Adminer provides web-based database management
- Direct PostgreSQL access available on port 5433
- Health check function available for monitoring

## Scalability Considerations

The current architecture supports several scalability patterns:

1. **Horizontal Scaling**: The stateless Go application can be scaled horizontally behind a load balancer
2. **Database Scaling**: PostgreSQL supports read replicas for scaling read operations
3. **Caching Layer**: Can be added between the application and database
4. **Message Queue**: Can be introduced for asynchronous job processing and notifications
5. **Microservices**: The monolithic application can be broken down into microservices as needed