# GigCo API Reference

Complete API documentation for the GigCo platform. All endpoints are prefixed with `/api/v1/` unless otherwise specified.

## Table of Contents
- [Authentication](#authentication)
- [Jobs](#jobs)
- [Payments](#payments)
- [Users & Workers](#users--workers)
- [Schedules](#schedules)
- [Reviews](#reviews)
- [Error Handling](#error-handling)

## Base URL
```
http://localhost:8080
```

## Authentication

### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword",
  "name": "John Doe",
  "role": "consumer"  // or "gig_worker"
}
```

**Response (201 Created):**
```json
{
  "id": 1,
  "uuid": "a1b2c3d4-...",
  "email": "user@example.com",
  "name": "John Doe",
  "role": "consumer",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response (200 OK):**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "role": "consumer"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## Jobs

### List Jobs
```http
GET /api/v1/jobs?status=posted&limit=20
```

**Query Parameters:**
- `status` (optional): Filter by job status (posted, accepted, in_progress, completed, etc.)
- `consumer_id` (optional): Filter by consumer ID
- `worker_id` (optional): Filter by worker ID
- `limit` (optional): Results per page (default: 20, max: 100)

**Response (200 OK):**
```json
{
  "jobs": [
    {
      "id": 1,
      "uuid": "job-uuid-...",
      "title": "Lawn Mowing Service",
      "description": "Need lawn mowed, about 1000 sq ft",
      "status": "posted",
      "price": 50.00,
      "currency": "USD",
      "consumer_id": 1,
      "consumer_name": "John Doe",
      "location_address": "123 Main St",
      "created_at": "2025-12-12T10:00:00Z"
    }
  ],
  "count": 1,
  "pagination": {
    "total": 1,
    "page": 1,
    "limit": 20
  }
}
```

### Get Available Jobs (Workers Only)
```http
GET /api/v1/jobs/available
Authorization: Bearer <token>
```

Returns jobs with `status=posted` that are available for workers to accept.

### Create Job (Consumers Only)
```http
POST /api/v1/jobs/create
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Lawn Mowing Service",
  "description": "Need lawn mowed, about 1000 sq ft",
  "price": 50.00,
  "currency": "USD",
  "location_address": "123 Main St",
  "location_city": "Springfield",
  "location_state": "IL",
  "location_zip": "62701",
  "scheduled_start": "2025-12-15T09:00:00Z",
  "estimated_duration_hours": 2
}
```

**Response (201 Created):**
```json
{
  "id": 1,
  "uuid": "job-uuid-...",
  "title": "Lawn Mowing Service",
  "status": "posted",
  "created_at": "2025-12-12T10:00:00Z"
}
```

### Accept Job (Workers Only)
```http
POST /api/v1/jobs/{id}/accept
Authorization: Bearer <token>
```

Accepts a job and triggers the Temporal workflow.

**Response (200 OK):**
```json
{
  "message": "Job accepted successfully",
  "job_id": 1,
  "status": "accepted",
  "workflow_id": "job-workflow-uuid-..."
}
```

### Start Job (Workers Only)
```http
POST /api/v1/jobs/{id}/start
Authorization: Bearer <token>
```

Marks job as in progress.

### Complete Job (Workers & Consumers)
```http
POST /api/v1/jobs/{id}/complete
Authorization: Bearer <token>
```

Dual confirmation system:
- Worker calls this to mark work complete
- Consumer calls this to confirm completion
- Job is fully completed when both confirm

**Response (200 OK):**
```json
{
  "message": "Job marked as completed by worker. Awaiting consumer confirmation.",
  "job_id": 1,
  "worker_completed_at": "2025-12-12T14:00:00Z",
  "consumer_completed_at": null
}
```

## Payments

### Authorize Payment (Consumers Only)
Pre-authorize payment and hold in escrow.

```http
POST /api/v1/payments/authorize
Authorization: Bearer <token>
Content-Type: application/json

{
  "job_id": 1,
  "amount": 50.00,
  "currency": "USD",
  "payment_method": "clover",
  "clover_source_token": "clv_token_..."
}
```

**Response (200 OK):**
```json
{
  "transaction_id": "txn-uuid-...",
  "job_id": 1,
  "amount": 50.00,
  "status": "authorized",
  "clover_charge_id": "clv_charge_...",
  "authorized_at": "2025-12-12T10:05:00Z"
}
```

### Capture Payment (Consumers & Workers)
Release payment from escrow after job completion.

```http
POST /api/v1/payments/capture
Authorization: Bearer <token>
Content-Type: application/json

{
  "transaction_id": "txn-uuid-...",
  "amount": 50.00  // optional, defaults to full authorized amount
}
```

**Response (200 OK):**
```json
{
  "transaction_id": "txn-uuid-...",
  "status": "captured",
  "amount": 50.00,
  "platform_fee": 5.00,
  "worker_payment": 45.00,
  "captured_at": "2025-12-12T15:00:00Z"
}
```

### Refund Payment (Consumers Only)
```http
POST /api/v1/payments/refund
Authorization: Bearer <token>
Content-Type: application/json

{
  "transaction_id": "txn-uuid-...",
  "amount": 50.00,  // optional for partial refund
  "reason": "Job cancelled"
}
```

**Response (200 OK):**
```json
{
  "transaction_id": "txn-uuid-...",
  "status": "refunded",
  "refund_amount": 50.00,
  "refunded_at": "2025-12-12T11:00:00Z"
}
```

### Get Payment Summary
```http
GET /api/v1/jobs/{id}/payment-summary
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "job_id": 1,
  "total_authorized": 50.00,
  "total_captured": 50.00,
  "total_refunded": 0.00,
  "platform_fees": 5.00,
  "worker_payment": 45.00,
  "escrow_status": "released"
}
```

### Get Job Transactions
```http
GET /api/v1/jobs/{id}/payments
Authorization: Bearer <token>
```

Returns all transactions for a specific job.

## Users & Workers

### Get User Profile
```http
GET /api/v1/users/profile
Authorization: Bearer <token>
```

Returns the authenticated user's profile.

### Register as Gig Worker
```http
POST /api/v1/gigworkers/create
Authorization: Bearer <token>
Content-Type: application/json

{
  "skills": ["lawn_care", "landscaping"],
  "hourly_rate": 25.00,
  "bio": "Experienced lawn care professional",
  "availability": "weekends"
}
```

### List Gig Workers
```http
GET /api/v1/gigworkers?skills=lawn_care&limit=20
```

**Query Parameters:**
- `skills` (optional): Filter by skills
- `min_rating` (optional): Minimum rating
- `limit` (optional): Results per page

## Schedules

### List Schedules
```http
GET /api/v1/schedules?worker_id=1&is_available=true&start_date=2025-12-15
```

**Query Parameters:**
- `worker_id` (optional): Filter by worker ID
- `is_available` (optional): Filter by availability (true/false)
- `start_date` (optional): Filter by start date (YYYY-MM-DD)
- `end_date` (optional): Filter by end date (YYYY-MM-DD)
- `limit` (optional): Results per page (default: 20, max: 100)

**Response (200 OK):**
```json
{
  "schedules": [
    {
      "id": 1,
      "uuid": "schedule-uuid-...",
      "gig_worker_id": 1,
      "title": "Weekend Availability",
      "start_time": "2025-12-15T08:00:00Z",
      "end_time": "2025-12-15T17:00:00Z",
      "is_available": true,
      "job_id": null
    }
  ],
  "count": 1
}
```

### Create Schedule
```http
POST /api/v1/schedules/create
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Weekend Availability",
  "start_time": "2025-12-15T08:00:00Z",
  "end_time": "2025-12-15T17:00:00Z",
  "is_available": true,
  "recurring_pattern": "weekly",
  "notes": "Available every Saturday"
}
```

## Reviews

### Submit Review
```http
POST /api/v1/jobs/{id}/review
Authorization: Bearer <token>
Content-Type: application/json

{
  "rating": 5,
  "comment": "Excellent service, very professional!",
  "reviewer_type": "consumer"  // or "gig_worker"
}
```

### Get Job Reviews
```http
GET /api/v1/jobs/{id}/reviews
```

### Get User Review Stats
```http
GET /api/v1/users/{id}/reviews
```

Returns aggregated review statistics for a user.

## Error Handling

All errors follow a consistent format:

**Error Response (4xx/5xx):**
```json
{
  "error": "Validation failed",
  "message": "Invalid input parameters",
  "code": "VALIDATION_ERROR",
  "details": {
    "amount": "must be greater than 0",
    "currency": "required field"
  }
}
```

### Common Error Codes
- `400 Bad Request`: Invalid input or validation error
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: Insufficient permissions for this action
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict (e.g., job already accepted)
- `500 Internal Server Error`: Server-side error

## Rate Limiting
Currently no rate limiting is implemented. In production, consider implementing rate limiting per user/IP.

## Pagination
All list endpoints support pagination:
- `limit`: Number of results per page (default: 20, max: 100)
- `offset`: Number of results to skip (for pagination)

## Testing

### Using cURL
```bash
# Health check
curl http://localhost:8080/health

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"consumer1@gigco.dev","password":"test123"}'

# Get jobs with authentication
curl http://localhost:8080/api/v1/jobs \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### Using Postman
Import the collection from `test/GigCo-API.postman_collection.json` for a complete testing suite.

---

**Need Help?** Check the main [README.md](./README.md) for setup instructions and development guidance.
