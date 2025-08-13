# GigCo API Testing with Postman

This directory contains Postman collections and environments for testing the GigCo API.

## Files

### Collections
- `GigCo-API.postman_collection.json` - Complete API test collection with automated tests

### Environments  
- `GigCo-Local.postman_environment.json` - Local Docker development environment variables

## Setup Instructions

### 1. Import Collection and Environment
1. Open Postman
2. Click **Import** button
3. Select both JSON files from this directory:
   - `GigCo-API.postman_collection.json`
   - `GigCo-Local.postman_environment.json`

### 2. Select Environment
1. In the top-right corner of Postman, select **GigCo Local Docker Environment**
2. Verify the `baseUrl` is set to `http://localhost:8080`

### 3. Start Docker Environment
Ensure your Docker containers are running:
```bash
docker compose up -d
```

## API Endpoints Included

### Health Check
- **GET** `/health`
- **Purpose**: Verify application and database connectivity
- **Tests**: Status code, healthy response, response time

### Customer Management  
- **GET** `/api/v1/customers/{id}`
- **Purpose**: Retrieve customer by ID
- **Tests**: Status code, customer data structure, ID matching

### User Creation
- **POST** `/api/v1/users/create`
- **Purpose**: Create new users with name and address
- **Tests**: Creation success, data validation, ID storage for chaining
- **Body**: 
  ```json
  {
    "name": "User Name",
    "address": "User Address"
  }
  ```

### User Registration
- **POST** `/api/v1/auth/register`
- **Purpose**: Register new users with role selection (consumer, gig_worker, admin)
- **Tests**: Registration success, email validation, role validation, duplicate email handling
- **Body**:
  ```json
  {
    "name": "Full Name",
    "email": "user@example.com",
    "address": "User Address",
    "role": "consumer",
    "phone": "+1234567890"
  }
  ```

### GigWorker Management
- **POST** `/api/v1/gigworkers/create`
- **GET** `/api/v1/gigworkers`
- **GET** `/api/v1/gigworkers/{id}`
- **Purpose**: Create and manage gig workers with detailed profiles
- **Tests**: Creation, validation, pagination, filtering by verification status
- **Create Body**:
  ```json
  {
    "name": "Worker Name",
    "email": "worker@example.com",
    "address": "Worker Address",
    "phone": "+1234567890",
    "hourly_rate": 25.00,
    "experience_years": 3,
    "service_radius_miles": 15.0,
    "bio": "Professional description"
  }
  ```

### Job Management
- **POST** `/api/v1/jobs/create`
- **GET** `/api/v1/jobs`
- **GET** `/api/v1/jobs/{id}`
- **POST** `/api/v1/jobs/{id}/accept`
- **Purpose**: Create, list, view, and accept jobs
- **Tests**: Job creation, pagination with filters, job acceptance, validation
- **Create Body**:
  ```json
  {
    "title": "Job Title",
    "description": "Detailed job description",
    "category": "cleaning",
    "location": "Job Location",
    "pay_rate": 25.00,
    "estimated_hours": 4,
    "consumer_id": 1
  }
  ```

### Schedule Management
- **POST** `/api/v1/schedules/create`
- **Purpose**: Create worker availability schedules
- **Tests**: Schedule creation, time validation, worker assignment
- **Body**:
  ```json
  {
    "gig_worker_id": 1,
    "title": "Availability Window",
    "start_time": "2025-08-15T09:00:00Z",
    "end_time": "2025-08-15T17:00:00Z",
    "is_available": true,
    "notes": "Available for cleaning jobs"
  }
  ```

### Transaction Management
- **POST** `/api/v1/transactions/create`
- **Purpose**: Create payment transactions
- **Tests**: Transaction creation, amount validation, required field validation
- **Body**:
  ```json
  {
    "job_id": 1,
    "consumer_id": 1,
    "gig_worker_id": 2,
    "amount": 100.00,
    "payment_method": "credit_card",
    "notes": "Payment for service"
  }
  ```

### Chained Test
- **GET** `/api/v1/customers/{lastCreatedUserId}`
- **Purpose**: Verify newly created user can be retrieved
- **Tests**: Automatic ID retrieval from previous test

## Test Features

### Automated Testing
- All requests include comprehensive test scripts
- Tests validate status codes, response structure, and data integrity
- Environment variables are automatically set and used between requests

### Dynamic Data
- Uses Postman's built-in dynamic variables (`{{$randomFullName}}`, `{{$randomStreetAddress}}`)
- Automatically chains requests (create user → get user)
- Stores created user IDs for follow-up tests

### Error Scenarios
- Includes sample responses for both success and error cases
- Tests handle missing required fields and not-found scenarios

## Running Tests

### Individual Requests
1. Select any request from the collection
2. Click **Send** to execute
3. View test results in the **Test Results** tab

### Collection Runner
1. Click **Collections** in the left sidebar
2. Hover over **GigCo API Collection**
3. Click **Run collection** (play button)
4. Select the **GigCo Local Docker Environment**
5. Click **Run GigCo API Collection**

### Expected Results
All tests should pass when Docker containers are running:
- ✅ Health Check: Application healthy
- ✅ Get Customer: Sample data retrieved  
- ✅ Create User: New user created successfully
- ✅ Get New Customer: Created user can be retrieved

## Sample Data

The database is seeded with test customers:
- ID 1: John Doe
- ID 2: Jane Smith  
- ID 3: Bob Johnson

## Troubleshooting

### Connection Refused
- Ensure Docker containers are running: `docker compose ps`
- Check port 8080 is not blocked: `curl http://localhost:8080/health`

### Tests Failing
- Verify environment is selected in Postman
- Check Docker logs: `docker compose logs app`
- Restart containers: `docker compose restart`

### Database Issues
- Check PostgreSQL health: `docker compose logs postgres`
- Connect directly: `docker compose exec postgres psql -U postgres -d gigco`

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `base_url` | `http://localhost:8080` | API base URL |
| `customerId` | `1` | Default customer for testing |
| `testCustomerId` | `2` | Alternative customer ID |
| `workerId` | `2` | Default gig worker ID for testing |
| `consumerId` | `1` | Default consumer ID for testing |
| `apiVersion` | `v1` | API version prefix |
| `lastCreatedUserId` | (dynamic) | Auto-set by create user tests |
| `lastCreatedGigWorkerId` | (dynamic) | Auto-set by create gigworker tests |
| `lastRegisteredUserId` | (dynamic) | Auto-set by registration tests |
| `testJobId` | (dynamic) | Auto-set by create job tests |
| `testScheduleId` | (dynamic) | Auto-set by create schedule tests |
| `testTransactionId` | (dynamic) | Auto-set by create transaction tests |

## Next Steps

As new API endpoints are added during development:
1. Add new requests to the collection
2. Include appropriate test scripts
3. Update environment variables as needed
4. Document any new testing procedures