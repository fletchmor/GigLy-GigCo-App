# Changelog

All notable changes to the GigCo platform are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-01-19

### 🚀 Production Ready Release

This release marks GigCo as production-ready with comprehensive security hardening, monitoring, and deployment capabilities.

### Added

#### Security Enhancements
- **Security Headers Middleware** (`internal/middleware/security.go`)
  - X-Content-Type-Options: nosniff
  - X-Frame-Options: DENY
  - X-XSS-Protection: 1; mode=block
  - Content-Security-Policy
  - Strict-Transport-Security (production)
  - Referrer-Policy
  - Permissions-Policy

- **CORS Protection** (`internal/middleware/security.go`)
  - Configurable allowed origins via `CORS_ALLOWED_ORIGINS`
  - Credentials support
  - Preflight request handling

- **Rate Limiting** (`internal/middleware/security.go`)
  - Standard: 100 requests/minute for API endpoints
  - Strict: 5 requests/minute for authentication endpoints
  - IP-based tracking with automatic cleanup

- **Strong Password Policy** (`api/auth.go`)
  - Minimum 10 characters
  - Maximum 72 characters (bcrypt limit)
  - Requires 3 of 4 character types (upper, lower, number, special)
  - Common password dictionary check

- **Admin Registration Protection** (`api/auth.go`)
  - Admin role registration blocked in production environment
  - Audit logging for admin registration attempts

- **iOS SSL Certificate Pinning** (`ios-app/.../URLSessionDelegate.swift`)
  - Public key pinning implementation
  - Support for certificate rotation (backup pins)
  - Automatic validation against pinned hashes

#### Infrastructure
- **Structured Logging** (`internal/logger/logger.go`)
  - zerolog integration for JSON structured logs
  - Configurable log levels via `LOG_LEVEL`
  - Request ID tracking
  - User context in logs

- **Error Tracking** (`internal/sentry/sentry.go`)
  - Sentry integration for production error monitoring
  - Automatic sensitive data scrubbing
  - Request context capture
  - Configurable via `SENTRY_DSN`

- **Health Check Endpoints** (`api/health.go`)
  - `GET /health` - Basic health check (backwards compatible)
  - `GET /ready` - Kubernetes readiness probe with dependency checks
  - `GET /live` - Kubernetes liveness probe
  - `GET /metrics` - Runtime metrics (memory, goroutines)

- **Graceful Shutdown** (`cmd/main.go`)
  - Signal handling (SIGTERM, SIGINT)
  - Connection draining with 30-second timeout
  - Clean database connection closure

- **HTTP Server Timeouts** (`cmd/main.go`)
  - ReadTimeout: 15 seconds
  - WriteTimeout: 15 seconds
  - IdleTimeout: 60 seconds

- **Production Configuration Validation** (`cmd/main.go`)
  - Required environment variables checked at startup
  - Fails fast if JWT_SECRET missing in production
  - Warnings for insecure configurations

#### Services
- **Email Service** (`internal/email/email.go`)
  - SendGrid integration
  - Email verification emails
  - Password reset emails
  - Job notification emails
  - HTML and plain text support

- **Push Notifications** (`internal/notifications/push.go`)
  - Firebase Cloud Messaging (FCM) integration
  - Device token management
  - Topic-based notifications
  - Job and payment notification helpers

#### DevOps
- **CI/CD Pipeline** (`.github/workflows/ci.yml`)
  - Linting: go vet, gofmt, staticcheck
  - Testing with PostgreSQL service container
  - Security scanning: gosec, govulncheck
  - Docker image build and push
  - Container vulnerability scanning (Trivy)
  - Code coverage reporting

- **Production Docker Configuration** (`docker-compose.prod.yml`)
  - No exposed database ports
  - No admin tools
  - Resource limits
  - Log rotation
  - Health checks

- **Load Testing** (`scripts/load_test.js`)
  - k6 load testing script
  - Configurable stages and thresholds
  - Authentication flow testing
  - API endpoint testing

#### Testing
- **Unit Tests** (`internal/auth/jwt_test.go`)
  - JWT generation and validation tests
  - Password hashing tests
  - Token generation tests
  - Benchmark tests

- **API Tests** (`api/auth_test.go`)
  - Password strength validation tests
  - Registration validation tests
  - Phone number normalization tests

#### Documentation
- **DEPLOYMENT.md** - Production deployment guide
- **SECURITY.md** - Security architecture documentation
- **CHANGELOG.md** - Version history (this file)
- Updated **README.md** with production status

### Changed

- **JWT Secret Handling** (`internal/auth/jwt.go`)
  - No longer logs generated secrets
  - Requires JWT_SECRET in production
  - Validates minimum secret length (32 chars)

- **Dockerfile**
  - Removed hardcoded database credentials
  - Added Docker HEALTHCHECK instruction
  - Only PORT environment variable set by default

- **iOS Configuration** (`ios-app/.../Configuration.swift`)
  - Development URL now configurable via `DEV_API_HOST`
  - Supports UserDefaults and environment variable override
  - Default changed from hardcoded IP to localhost

- **.env.production.template**
  - Added `APP_ENV` variable
  - Changed `ALLOWED_ORIGINS` to `CORS_ALLOWED_ORIGINS`
  - Added comprehensive configuration checklist

### Security

- Removed all hardcoded credentials from codebase
- JWT secrets no longer logged
- Production requires encrypted database connections (DB_SSLMODE=require)
- Admin registration blocked via public API in production
- Rate limiting on all endpoints
- Security headers on all responses

### Dependencies Added

- `github.com/rs/zerolog` - Structured logging
- `github.com/getsentry/sentry-go` - Error tracking

---

## [0.9.0] - 2025-12-12

### Added
- Clover payment integration (authorize/capture/refund)
- Payment escrow system
- Job payment summary endpoint
- Transaction tracking

### Changed
- Updated job workflow with payment integration

---

## [0.8.0] - 2025-10-09

### Added
- iOS mobile application (SwiftUI)
- Dual completion system (worker + consumer confirmation)
- Username display on job listings
- Job workflow endpoints (start, complete, reject)

### Fixed
- iOS navigation - jobs now tappable
- Authentication handling for NULL passwords
- Job status enum expansion

---

## [0.7.0] - 2025-09-15

### Added
- JWT authentication system
- Role-based access control (RBAC)
- Protected API endpoints
- User profile management

---

## [0.6.0] - 2025-08-20

### Added
- Temporal workflow integration
- Job acceptance workflow
- Workflow monitoring UI
- Worker service

---

## [0.5.0] - 2025-07-10

### Added
- Review and rating system
- Notification system
- Worker templates and services
- Schedule management

---

## [0.4.0] - 2025-06-01

### Added
- GigWorker management
- Worker profiles
- Service categories
- Location-based features

---

## [0.3.0] - 2025-05-01

### Added
- Job management system
- Job status workflow
- Consumer-worker matching

---

## [0.2.0] - 2025-04-01

### Added
- User registration and authentication
- Role system (consumer, gig_worker, admin)
- Basic API structure

---

## [0.1.0] - 2025-03-01

### Added
- Initial project setup
- Docker development environment
- PostgreSQL database schema
- Basic health check endpoint
- Chi router configuration

---

## Migration Notes

### Upgrading to 1.0.0

1. **Environment Variables**: Add new required variables:
   ```bash
   APP_ENV=production
   CORS_ALLOWED_ORIGINS=https://your-domain.com
   ```

2. **Password Policy**: Existing users with weak passwords will still work, but new passwords must meet the new requirements.

3. **iOS App**: Update certificate pins in `URLSessionDelegate.swift` with your production certificate hashes.

4. **Database**: No schema changes required.

5. **Docker**: If using docker-compose, switch to `docker-compose.prod.yml` for production.

---

**Repository**: https://github.com/your-org/gigco
**Documentation**: See README.md, DEPLOYMENT.md, SECURITY.md
