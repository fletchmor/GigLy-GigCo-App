# GigCo Production Deployment Guide

This guide covers deploying GigCo to a production environment.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Environment Setup](#environment-setup)
- [Database Setup](#database-setup)
- [SSL/TLS Configuration](#ssltls-configuration)
- [Docker Deployment](#docker-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Health Checks](#health-checks)
- [Monitoring & Logging](#monitoring--logging)
- [Scaling](#scaling)
- [Backup & Recovery](#backup--recovery)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Services
- **PostgreSQL 17+**: Managed database (AWS RDS, Google Cloud SQL, or DigitalOcean)
- **Docker**: For containerized deployment
- **SSL Certificate**: From Let's Encrypt or commercial CA
- **Domain Name**: Configured with DNS pointing to your server

### Optional Services
- **SendGrid**: For transactional emails
- **Sentry**: For error tracking
- **Firebase**: For push notifications
- **Clover**: For payment processing

### System Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 1 vCPU | 2+ vCPU |
| Memory | 512 MB | 1+ GB |
| Storage | 10 GB | 20+ GB |
| Database | 1 GB | 5+ GB |

## Environment Setup

### 1. Create Production Environment File

```bash
# Copy the template
cp .env.production.template .env.production

# Edit with your production values
nano .env.production
```

### 2. Required Environment Variables

```bash
# Application
APP_ENV=production
PORT=8080
APP_VERSION=1.0.0
APP_BASE_URL=https://app.gigco.com

# Database (use managed database service)
DB_HOST=your-db-host.rds.amazonaws.com
DB_PORT=5432
DB_NAME=gigco_production
DB_USER=gigco_app
DB_PASSWORD=<strong-password-here>
DB_SSLMODE=require

# Security
JWT_SECRET=<generate-with-openssl-rand-base64-64>

# CORS (comma-separated list of allowed origins)
CORS_ALLOWED_ORIGINS=https://app.gigco.com,https://www.gigco.com

# TLS Certificates
TLS_CERT=/app/certs/fullchain.pem
TLS_KEY=/app/certs/privkey.pem
```

### 3. Optional Service Configuration

```bash
# Email (SendGrid)
SENDGRID_API_KEY=SG.xxxx
EMAIL_FROM=noreply@gigco.com
EMAIL_FROM_NAME=GigCo

# Error Tracking (Sentry)
SENTRY_DSN=https://xxxx@sentry.io/xxxx

# Push Notifications (Firebase)
FCM_SERVER_KEY=xxxx
FIREBASE_PROJECT_ID=gigco-app

# Payments (Clover)
CLOVER_ENVIRONMENT=production
CLOVER_MERCHANT_ID=xxxx
CLOVER_ACCESS_TOKEN=xxxx
CLOVER_API_ACCESS_KEY=xxxx

# Logging
LOG_LEVEL=info
```

### 4. Generate Secure JWT Secret

```bash
# Generate a 64-character random secret
openssl rand -base64 64

# Or use this command
head -c 64 /dev/urandom | base64
```

## Database Setup

### 1. Create Production Database

```sql
-- Connect to PostgreSQL as admin
CREATE DATABASE gigco_production;
CREATE USER gigco_app WITH ENCRYPTED PASSWORD 'your-secure-password';
GRANT ALL PRIVILEGES ON DATABASE gigco_production TO gigco_app;

-- Enable required extensions
\c gigco_production
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

### 2. Run Migrations

```bash
# Apply the schema
psql -h your-db-host -U gigco_app -d gigco_production -f scripts/init.sql
```

### 3. Database Connection Pooling

For high-traffic deployments, consider using PgBouncer:

```bash
# Install PgBouncer
apt-get install pgbouncer

# Configure /etc/pgbouncer/pgbouncer.ini
[databases]
gigco_production = host=your-db-host port=5432 dbname=gigco_production

[pgbouncer]
listen_addr = 127.0.0.1
listen_port = 6432
pool_mode = transaction
max_client_conn = 1000
default_pool_size = 25
```

## SSL/TLS Configuration

### Option 1: Let's Encrypt (Recommended)

```bash
# Install certbot
apt-get install certbot

# Obtain certificate
certbot certonly --standalone -d api.gigco.com

# Certificates will be at:
# /etc/letsencrypt/live/api.gigco.com/fullchain.pem
# /etc/letsencrypt/live/api.gigco.com/privkey.pem

# Set up auto-renewal
certbot renew --dry-run
```

### Option 2: Commercial Certificate

Place your certificate files:
```
certs/
├── fullchain.pem    # Certificate + intermediate chain
└── privkey.pem      # Private key
```

### Configure in Environment

```bash
TLS_CERT=/app/certs/fullchain.pem
TLS_KEY=/app/certs/privkey.pem
```

## Docker Deployment

### 1. Build Production Image

```bash
# Build the image
docker build -t gigco-api:latest .

# Tag for registry
docker tag gigco-api:latest ghcr.io/your-org/gigco-api:latest
docker tag gigco-api:latest ghcr.io/your-org/gigco-api:v1.0.0

# Push to registry
docker push ghcr.io/your-org/gigco-api:latest
docker push ghcr.io/your-org/gigco-api:v1.0.0
```

### 2. Deploy with Docker Compose

```bash
# Deploy using production compose file
docker-compose -f docker-compose.prod.yml --env-file .env.production up -d

# Check status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f app
```

### 3. Production Docker Compose Structure

The `docker-compose.prod.yml` file includes:
- No exposed database ports (security)
- No admin tools (Adminer, Temporal UI)
- Resource limits
- Health checks
- Proper restart policies
- Log rotation

## Kubernetes Deployment

### 1. Create Namespace

```bash
kubectl create namespace gigco-production
```

### 2. Create Secrets

```bash
kubectl create secret generic gigco-secrets \
  --namespace gigco-production \
  --from-literal=db-password='your-db-password' \
  --from-literal=jwt-secret='your-jwt-secret' \
  --from-literal=sendgrid-api-key='your-sendgrid-key' \
  --from-literal=sentry-dsn='your-sentry-dsn'
```

### 3. Sample Kubernetes Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gigco-api
  namespace: gigco-production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: gigco-api
  template:
    metadata:
      labels:
        app: gigco-api
    spec:
      containers:
      - name: gigco-api
        image: ghcr.io/your-org/gigco-api:v1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: APP_ENV
          value: "production"
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: gigco-secrets
              key: db-password
        resources:
          limits:
            cpu: "1"
            memory: "512Mi"
          requests:
            cpu: "500m"
            memory: "256Mi"
        livenessProbe:
          httpGet:
            path: /live
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### 4. Apply Configuration

```bash
kubectl apply -f k8s/
```

## Health Checks

### Available Endpoints

| Endpoint | Purpose | Use Case |
|----------|---------|----------|
| `GET /health` | Basic health check | Load balancer health checks |
| `GET /ready` | Readiness check | K8s readiness probe, includes DB check |
| `GET /live` | Liveness check | K8s liveness probe |
| `GET /metrics` | Runtime metrics | Monitoring systems |

### Health Check Examples

```bash
# Basic health check
curl https://api.gigco.com/health
# Response: {"status":"healthy","timestamp":"..."}

# Readiness check (includes dependencies)
curl https://api.gigco.com/ready
# Response: {"status":"healthy","checks":{"database":{"status":"healthy"}}}

# Liveness check
curl https://api.gigco.com/live
# Response: {"status":"alive","uptime":"2h30m15s"}

# Metrics
curl https://api.gigco.com/metrics
# Response: {"runtime":{"goroutines":15},"memory":{"alloc_mb":25}}
```

## Monitoring & Logging

### Structured Logging

The application uses zerolog for structured JSON logging:

```json
{
  "level": "info",
  "service": "gigco-api",
  "time": "2026-01-19T10:30:00Z",
  "caller": "api/auth.go:125",
  "message": "User login successful",
  "user_id": 123,
  "request_id": "abc-123"
}
```

### Log Levels

Set via `LOG_LEVEL` environment variable:
- `debug`: Verbose debugging information
- `info`: General operational information (default)
- `warn`: Warning messages
- `error`: Error messages only

### Sentry Integration

Error tracking is automatic when `SENTRY_DSN` is configured:

```bash
SENTRY_DSN=https://xxxx@sentry.io/xxxx
```

Features:
- Automatic error capture
- Request context included
- Sensitive data scrubbed automatically
- Source maps for stack traces

### Metrics Collection

For production monitoring, integrate with:
- **Prometheus**: Scrape `/metrics` endpoint
- **Datadog**: Use Datadog agent
- **New Relic**: Use New Relic Go agent

## Scaling

### Horizontal Scaling

The API is stateless and can be horizontally scaled:

```bash
# Docker Compose
docker-compose -f docker-compose.prod.yml up -d --scale app=3

# Kubernetes
kubectl scale deployment gigco-api --replicas=5 -n gigco-production
```

### Load Balancing

Use a load balancer in front of multiple instances:
- AWS Application Load Balancer
- Google Cloud Load Balancing
- nginx/HAProxy

### Database Scaling

For database performance:
1. **Connection Pooling**: Use PgBouncer
2. **Read Replicas**: For read-heavy workloads
3. **Vertical Scaling**: Upgrade instance size

## Backup & Recovery

### Database Backups

```bash
# Manual backup
pg_dump -h your-db-host -U gigco_app -d gigco_production > backup_$(date +%Y%m%d).sql

# Restore from backup
psql -h your-db-host -U gigco_app -d gigco_production < backup_20260119.sql
```

### Automated Backups

For managed databases:
- **AWS RDS**: Enable automated backups
- **Google Cloud SQL**: Configure backup schedule
- **DigitalOcean**: Enable daily backups

### Disaster Recovery

1. **Point-in-time Recovery**: Use database transaction logs
2. **Cross-region Replication**: For critical deployments
3. **Regular Restore Testing**: Verify backups work

## Troubleshooting

### Common Issues

#### Application Won't Start

```bash
# Check logs
docker-compose -f docker-compose.prod.yml logs app

# Common causes:
# - Missing environment variables
# - Database connection failed
# - Port already in use
```

#### Database Connection Issues

```bash
# Test database connectivity
psql -h your-db-host -U gigco_app -d gigco_production -c "SELECT 1"

# Check if SSL is required
# Ensure DB_SSLMODE=require is set
```

#### Health Check Failing

```bash
# Check readiness endpoint
curl -v https://api.gigco.com/ready

# If database check fails, verify:
# 1. Database is running
# 2. Credentials are correct
# 3. Network connectivity exists
```

### Debug Mode

For troubleshooting, temporarily enable debug logging:

```bash
LOG_LEVEL=debug docker-compose -f docker-compose.prod.yml up
```

**Warning**: Do not run debug logging in production for extended periods.

### Getting Help

1. Check application logs
2. Review [SECURITY.md](./SECURITY.md) for security-related issues
3. Check [API_REFERENCE.md](./API_REFERENCE.md) for API issues
4. Review GitHub Issues for known problems

---

## Quick Deployment Checklist

- [ ] Managed PostgreSQL database created
- [ ] SSL certificate obtained
- [ ] `.env.production` configured with all required variables
- [ ] JWT_SECRET is 64+ characters
- [ ] DB_SSLMODE=require
- [ ] CORS_ALLOWED_ORIGINS set to your domains only
- [ ] Health checks verified
- [ ] Monitoring configured (Sentry, metrics)
- [ ] Backups configured
- [ ] Load testing completed

---

**Document Version**: 1.0.0
**Last Updated**: January 19, 2026
