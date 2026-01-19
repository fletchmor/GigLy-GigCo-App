# GigCo Security Documentation

This document outlines the security architecture, configurations, and best practices implemented in GigCo.

## Table of Contents

- [Security Overview](#security-overview)
- [Authentication](#authentication)
- [Authorization](#authorization)
- [API Security](#api-security)
- [Data Protection](#data-protection)
- [Mobile App Security](#mobile-app-security)
- [Infrastructure Security](#infrastructure-security)
- [Security Configuration](#security-configuration)
- [Vulnerability Reporting](#vulnerability-reporting)

## Security Overview

GigCo implements defense-in-depth security with multiple layers:

```
┌─────────────────────────────────────────────────────────────┐
│                      Client Layer                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  iOS App    │  │  Web App    │  │  API Client │        │
│  │ SSL Pinning │  │   HTTPS     │  │   HTTPS     │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Transport Layer                          │
│           TLS 1.2+ with Strong Cipher Suites               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Application Layer                         │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐      │
│  │  CORS    │ │  Rate    │ │ Security │ │   JWT    │      │
│  │ Headers  │ │ Limiting │ │ Headers  │ │   Auth   │      │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Data Layer                              │
│     Encrypted at Rest │ Parameterized Queries │ RBAC       │
└─────────────────────────────────────────────────────────────┘
```

## Authentication

### JWT Token Authentication

GigCo uses JSON Web Tokens (JWT) for stateless authentication.

#### Token Structure

```json
{
  "user_id": 123,
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "role": "consumer",
  "exp": 1705660800,
  "iat": 1705574400,
  "iss": "gigco-api"
}
```

#### Token Lifecycle

| Stage | Duration | Action |
|-------|----------|--------|
| Issue | Login | Token generated with 24-hour expiry |
| Refresh | < 1 hour remaining | New token issued |
| Expiry | 24 hours | Token invalid, re-login required |

#### Security Measures

- **Algorithm**: HS256 (HMAC-SHA256)
- **Secret**: Minimum 32 characters, required in production
- **Validation**: Signature, expiry, issuer verified on every request

### Password Security

#### Requirements

Passwords must meet ALL of the following:

| Requirement | Value |
|-------------|-------|
| Minimum Length | 10 characters |
| Maximum Length | 72 characters (bcrypt limit) |
| Complexity | At least 3 of: uppercase, lowercase, numbers, special chars |
| Common Passwords | Blocked (dictionary check) |

#### Storage

- **Algorithm**: bcrypt
- **Cost Factor**: 10 (default)
- **Format**: `$2a$10$...` (60 characters)

#### Implementation

```go
// Password validation (api/auth.go)
func validatePasswordStrength(password string) error {
    // Length check
    // Complexity check (3 of 4 character types)
    // Common password check
}
```

### Session Security

- Tokens transmitted only via HTTPS
- Tokens stored in mobile Keychain (iOS)
- No tokens stored in cookies or localStorage
- Logout invalidates client-side token

## Authorization

### Role-Based Access Control (RBAC)

#### Roles

| Role | Description | Capabilities |
|------|-------------|--------------|
| `consumer` | Service requesters | Post jobs, manage own jobs, make payments |
| `gig_worker` | Service providers | Accept jobs, complete jobs, receive payments |
| `admin` | System administrators | Full access, user management |

#### Route Protection

```go
// Protected routes require authentication
r.Group(func(r chi.Router) {
    r.Use(middleware.JWTAuth)

    // Role-specific routes
    r.With(middleware.RequireRole("admin")).Post("/api/v1/users/create", ...)
    r.With(middleware.RequireRoles("admin", "consumer")).Post("/api/v1/jobs/create", ...)
})
```

#### Admin Registration Protection

Admin accounts cannot be created via the public API in production:

```go
if req.Role == "admin" && os.Getenv("APP_ENV") == "production" {
    return fmt.Errorf("admin registration is not allowed via public API")
}
```

## API Security

### Security Headers

Applied to all responses via `SecurityHeaders` middleware:

| Header | Value | Purpose |
|--------|-------|---------|
| `X-Content-Type-Options` | `nosniff` | Prevent MIME-type sniffing |
| `X-Frame-Options` | `DENY` | Prevent clickjacking |
| `X-XSS-Protection` | `1; mode=block` | XSS filter (legacy browsers) |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Control referrer information |
| `Content-Security-Policy` | `default-src 'self'; frame-ancestors 'none'` | Restrict resource loading |
| `Permissions-Policy` | `geolocation=(), microphone=(), camera=()` | Disable unused features |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` | Force HTTPS (production only) |

### CORS Configuration

Cross-Origin Resource Sharing is strictly controlled:

```bash
# Environment variable
CORS_ALLOWED_ORIGINS=https://app.gigco.com,https://www.gigco.com
```

**Important**: Never use `*` as allowed origin in production.

### Rate Limiting

Protection against brute force and DoS attacks:

| Endpoint Type | Rate | Window |
|---------------|------|--------|
| Standard API | 100 requests | 1 minute |
| Authentication | 5 requests | 1 minute |

Implementation:
- IP-based tracking
- Automatic cleanup of stale entries
- `429 Too Many Requests` response with `Retry-After` header

### Input Validation

All user input is validated:

- **Email**: Regex validation
- **Phone**: Format validation with normalization
- **Passwords**: Strength validation
- **Names**: Length limits (2-255 characters)
- **SQL Queries**: Parameterized (no string concatenation)

## Data Protection

### Database Security

#### Connection Security

```bash
# Production database connection
DB_SSLMODE=require  # Encrypted connections only
```

#### Access Control

- Application uses dedicated database user
- Least-privilege principle (only required permissions)
- No shared database credentials

#### Sensitive Data Handling

| Data Type | Storage | Access |
|-----------|---------|--------|
| Passwords | bcrypt hash | Never retrievable |
| JWT Secret | Environment variable | Server only |
| API Keys | Environment variable | Server only |
| Payment Tokens | Clover (tokenized) | Never stored locally |

### Logging Security

Sensitive data is never logged:

```go
// Scrubbed from logs:
- Passwords
- JWT tokens
- API keys
- Credit card numbers
- Personal identifiers
```

### Error Handling

Production error responses:
- Generic error messages to clients
- Detailed errors logged server-side
- Stack traces only in development

## Mobile App Security

### iOS Security Features

#### Keychain Storage

Tokens stored securely in iOS Keychain:

```swift
// KeychainHelper.swift
kSecAttrAccessible: kSecAttrAccessibleAfterFirstUnlock
```

#### SSL Certificate Pinning

Public key pinning prevents MITM attacks:

```swift
// URLSessionDelegate.swift
class SecureURLSessionDelegate: NSObject, URLSessionDelegate {
    private static let pinnedPublicKeyHashes: [String] = [
        "AAAAAAA...",  // Primary certificate hash
        "BBBBBBB...",  // Backup certificate hash
    ]

    func validateCertificateChain(serverTrust: SecTrust) -> Bool {
        // Validates certificate against pinned hashes
    }
}
```

#### Generating Certificate Hashes

```bash
# 1. Get certificate
openssl s_client -connect api.gigco.com:443 </dev/null 2>/dev/null | \
  openssl x509 -outform DER > cert.der

# 2. Extract public key
openssl x509 -inform DER -in cert.der -pubkey -noout > pubkey.pem

# 3. Get SHA256 hash
openssl pkey -pubin -in pubkey.pem -outform DER | \
  openssl dgst -sha256 -binary | base64
```

#### Debug vs Release Builds

| Feature | Debug | Release |
|---------|-------|---------|
| SSL Validation | Relaxed (self-signed OK) | Strict (pinning enabled) |
| Logging | Verbose | Minimal |
| Error Details | Full | Generic |

## Infrastructure Security

### Docker Security

Production Docker configuration:

```dockerfile
# Non-root user (if applicable)
# No hardcoded secrets
# Minimal base image (alpine)
# Health checks enabled
```

### Network Security

```yaml
# docker-compose.prod.yml
# - Database port NOT exposed externally
# - Internal Docker network for service communication
# - Only API port (8080) exposed
```

### Secrets Management

**Never commit secrets to version control.**

Secrets should be managed via:
- Environment variables
- Docker secrets
- Kubernetes secrets
- AWS Secrets Manager / HashiCorp Vault

### CI/CD Security

GitHub Actions pipeline includes:

1. **Static Analysis**: `go vet`, `staticcheck`
2. **Security Scanning**: `gosec` (SAST)
3. **Dependency Scanning**: `govulncheck`
4. **Container Scanning**: Trivy

## Security Configuration

### Environment Variables Reference

| Variable | Required | Description |
|----------|----------|-------------|
| `APP_ENV` | Yes | Must be `production` in production |
| `JWT_SECRET` | Yes | Min 32 chars, randomly generated |
| `DB_PASSWORD` | Yes | Strong password |
| `DB_SSLMODE` | Yes | Must be `require` in production |
| `CORS_ALLOWED_ORIGINS` | Yes | Comma-separated allowed origins |
| `SENTRY_DSN` | No | Error tracking endpoint |

### Security Checklist

Before deploying to production:

- [ ] `APP_ENV=production` is set
- [ ] `JWT_SECRET` is 64+ characters, randomly generated
- [ ] `DB_SSLMODE=require` is set
- [ ] `CORS_ALLOWED_ORIGINS` contains only your domains
- [ ] SSL certificates are valid and not self-signed
- [ ] Database credentials are unique and strong
- [ ] Admin accounts created via database, not API
- [ ] Rate limiting is enabled
- [ ] Security headers are verified
- [ ] Logging does not contain sensitive data
- [ ] iOS certificate pins are updated with production certs

### Verifying Security Configuration

```bash
# Check security headers
curl -I https://api.gigco.com/health

# Expected headers:
# X-Content-Type-Options: nosniff
# X-Frame-Options: DENY
# Strict-Transport-Security: max-age=31536000; includeSubDomains

# Check CORS (should fail from unauthorized origin)
curl -H "Origin: https://malicious.com" \
     -H "Access-Control-Request-Method: GET" \
     -X OPTIONS https://api.gigco.com/api/v1/jobs

# Verify SSL/TLS
openssl s_client -connect api.gigco.com:443 -tls1_2
```

## Vulnerability Reporting

### Responsible Disclosure

If you discover a security vulnerability:

1. **Do NOT** disclose publicly
2. Email security@gigco.com with:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
3. Allow 90 days for fix before disclosure

### Bug Bounty

Currently, no formal bug bounty program exists. We appreciate security researchers who report issues responsibly.

### Security Updates

Security patches are released as soon as possible after discovery. Subscribe to releases for notifications.

---

## Appendix: Security File Locations

| File | Purpose |
|------|---------|
| `internal/middleware/security.go` | Security headers, CORS, rate limiting |
| `internal/auth/jwt.go` | JWT generation and validation |
| `api/auth.go` | Password validation, authentication |
| `ios-app/.../URLSessionDelegate.swift` | iOS SSL pinning |
| `ios-app/.../KeychainHelper.swift` | iOS secure storage |
| `.github/workflows/ci.yml` | Security scanning in CI |

---

**Document Version**: 1.0.0
**Last Updated**: January 19, 2026
