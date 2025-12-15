# 🚀 Quick Start: Production in 1 Week

This guide focuses on the **bare minimum** to launch safely in production.

---

## Day 1-2: Security Basics

### 1. Set Strong JWT Secret
```bash
# Generate a strong secret
openssl rand -base64 64

# Add to production .env
JWT_SECRET=your-generated-secret-here
```

### 2. Enable HTTPS with Nginx
```nginx
# /etc/nginx/sites-available/gigco
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name api.yourdomain.com;
    return 301 https://$server_name$request_uri;
}
```

**Install SSL Certificate:**
```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d api.yourdomain.com
```

### 3. Add Rate Limiting Middleware

Create `/Users/fletcher/app/internal/middleware/ratelimit.go`:
```go
package middleware

import (
    "net/http"
    "sync"
    "time"
    "golang.org/x/time/rate"
)

var limiters = make(map[string]*rate.Limiter)
var mu sync.Mutex

func RateLimit(requestsPerSecond int) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ip := r.RemoteAddr

            mu.Lock()
            limiter, exists := limiters[ip]
            if !exists {
                limiter = rate.NewLimiter(rate.Limit(requestsPerSecond), requestsPerSecond*2)
                limiters[ip] = limiter
            }
            mu.Unlock()

            if !limiter.Allow() {
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

Add to `cmd/main.go`:
```go
NewServer.Use(middleware.RateLimit(10)) // 10 requests per second
```

---

## Day 3: Database Production Setup

### 1. Use Managed Database (AWS RDS)
```bash
# Don't use Docker PostgreSQL in production!
# Set up AWS RDS PostgreSQL or DigitalOcean Managed Database
```

### 2. Update Environment Variables
```bash
DB_HOST=your-rds-endpoint.amazonaws.com
DB_SSLMODE=require  # Force SSL!
DB_PASSWORD=strong-random-password
```

### 3. Create Database Indexes
```sql
-- Connect to production database
psql -h your-db-host -U gigco_user -d gigco

-- Add critical indexes
CREATE INDEX CONCURRENTLY idx_jobs_status ON jobs(status);
CREATE INDEX CONCURRENTLY idx_jobs_consumer_id ON jobs(consumer_id);
CREATE INDEX CONCURRENTLY idx_jobs_created_at ON jobs(created_at DESC);
CREATE INDEX CONCURRENTLY idx_gigworkers_user_id ON gigworkers(user_id);
CREATE INDEX CONCURRENTLY idx_people_email ON people(email);
```

### 4. Enable Automated Backups
In AWS RDS:
- Enable automated backups (7-day retention minimum)
- Enable point-in-time recovery
- Set backup window to low-traffic hours

---

## Day 4: Monitoring & Logging

### 1. Set Up Error Tracking (Sentry)
```bash
go get github.com/getsentry/sentry-go
```

Create `internal/monitoring/sentry.go`:
```go
package monitoring

import (
    "log"
    "os"
    "github.com/getsentry/sentry-go"
)

func InitSentry() {
    err := sentry.Init(sentry.ClientOptions{
        Dsn: os.Getenv("SENTRY_DSN"),
        Environment: os.Getenv("ENV"),
    })
    if err != nil {
        log.Printf("Sentry initialization failed: %v", err)
    }
}

func CaptureError(err error, context map[string]interface{}) {
    sentry.WithScope(func(scope *sentry.Scope) {
        scope.SetContext("additional", context)
        sentry.CaptureException(err)
    })
}
```

Add to `cmd/main.go`:
```go
import "app/internal/monitoring"

func main() {
    monitoring.InitSentry()
    defer sentry.Flush(2 * time.Second)
    // ... rest of code
}
```

Use in handlers:
```go
if err != nil {
    log.Printf("Database error: %v", err)
    monitoring.CaptureError(err, map[string]interface{}{
        "job_id": jobID,
        "user_id": userID,
    })
    http.Error(w, "Internal server error", 500)
    return
}
```

### 2. Improve Logging
```go
// Add structured logging
import "log"

// In handlers, log with context
log.Printf("[ERROR] Failed to fetch job: jobID=%d, userID=%d, error=%v",
    jobID, userID, err)

log.Printf("[INFO] User logged in: userID=%d, email=%s",
    user.ID, user.Email)
```

---

## Day 5: Configuration & Environment

### 1. Create Production .env
```bash
# .env.production (NEVER commit this!)
ENV=production

# Database (Managed Service)
DB_HOST=gigco-prod.xxxx.us-east-1.rds.amazonaws.com
DB_PORT=5432
DB_NAME=gigco_production
DB_USER=gigco_prod_user
DB_PASSWORD=<STRONG_PASSWORD_HERE>
DB_SSLMODE=require

# JWT (Use strong secret!)
JWT_SECRET=<64_CHARACTER_RANDOM_STRING>

# Clover (Production credentials)
CLOVER_ENVIRONMENT=production
CLOVER_MERCHANT_ID=<YOUR_MERCHANT_ID>
CLOVER_ACCESS_TOKEN=<YOUR_ACCESS_TOKEN>
CLOVER_API_ACCESS_KEY=<YOUR_API_KEY>

# Server
PORT=8080
ALLOWED_ORIGINS=https://yourdomain.com

# Temporal
TEMPORAL_HOST=temporal-production:7233

# Monitoring
SENTRY_DSN=<YOUR_SENTRY_DSN>
LOG_LEVEL=info
```

### 2. Secure Environment Variables
```bash
# NEVER commit .env files!
echo ".env*" >> .gitignore
echo "!.env.example" >> .gitignore

# Verify
git status
```

---

## Day 6: Deployment

### Option A: Simple VPS (DigitalOcean/Linode)
```bash
# 1. Create VPS ($12-24/month)
# 2. Install Docker and Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 3. Clone your repo
git clone https://github.com/yourusername/gigco.git
cd gigco

# 4. Create production .env file
nano .env

# 5. Start services
docker-compose up -d

# 6. Verify
docker-compose ps
curl http://localhost:8080/health
```

### Option B: AWS ECS (More Scalable)
1. Push Docker image to ECR
2. Create ECS cluster
3. Create task definition
4. Configure load balancer
5. Deploy service

---

## Day 7: Testing & Launch

### 1. Smoke Tests
```bash
# Test health endpoint
curl https://api.yourdomain.com/health

# Test login
curl -X POST https://api.yourdomain.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}'

# Test authenticated endpoint
TOKEN="your-token-here"
curl https://api.yourdomain.com/api/v1/jobs \
  -H "Authorization: Bearer $TOKEN"
```

### 2. Load Test (Optional but Recommended)
```bash
# Install k6
brew install k6  # Mac
# or
sudo apt install k6  # Linux

# Create load test script
cat > load_test.js << 'EOF'
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 10 },  // Ramp up to 10 users
    { duration: '3m', target: 10 },  // Stay at 10 users
    { duration: '1m', target: 0 },   // Ramp down
  ],
};

export default function () {
  const res = http.get('https://api.yourdomain.com/health');
  check(res, { 'status is 200': (r) => r.status === 200 });
  sleep(1);
}
EOF

# Run load test
k6 run load_test.js
```

### 3. Launch Checklist
- [ ] SSL certificate valid
- [ ] Database backups enabled
- [ ] Sentry error tracking working
- [ ] All environment variables set
- [ ] Docker containers running
- [ ] Health check returns 200
- [ ] Can create account
- [ ] Can login
- [ ] Can create job
- [ ] Can accept job
- [ ] Payment endpoints accessible
- [ ] Monitoring dashboard open

---

## 🆘 Common Issues & Solutions

### Issue: "Connection refused" on port 8080
```bash
# Check if app is running
docker-compose ps

# Check logs
docker-compose logs app

# Restart
docker-compose restart app
```

### Issue: "Database connection failed"
```bash
# Verify database is accessible
psql -h $DB_HOST -U $DB_USER -d $DB_NAME

# Check environment variables
docker-compose exec app env | grep DB_

# Test from container
docker-compose exec app ping $DB_HOST
```

### Issue: JWT tokens not working
```bash
# Verify JWT_SECRET is set
docker-compose exec app env | grep JWT_SECRET

# Check logs for JWT errors
docker-compose logs app | grep JWT
```

---

## 📞 Post-Launch Monitoring

### First 24 Hours:
- [ ] Check error rate every 2 hours
- [ ] Monitor response times
- [ ] Watch database connections
- [ ] Review Sentry errors
- [ ] Check disk space

### First Week:
- [ ] Review error patterns daily
- [ ] Optimize slow queries
- [ ] Fix critical bugs
- [ ] Gather user feedback
- [ ] Plan improvements

---

## 💰 Recommended Minimal Setup

**Total: ~$50-80/month**

1. **VPS**: DigitalOcean Droplet ($24/month)
   - 2 vCPUs, 4GB RAM

2. **Database**: Managed PostgreSQL ($15/month)
   - Auto-backups included

3. **Domain**: Namecheap (~$12/year)

4. **SSL**: Let's Encrypt (Free)

5. **Monitoring**: Sentry free tier (Free)

6. **CDN**: Cloudflare free tier (Free)

---

## 🎯 Success Criteria

Before you launch, you should be able to:
- ✅ Access API over HTTPS
- ✅ Create new user accounts
- ✅ Login and receive JWT token
- ✅ Create, view, and accept jobs
- ✅ Process payments (test mode)
- ✅ See errors in Sentry
- ✅ Automatic database backups
- ✅ Rate limiting prevents abuse
- ✅ SSL certificate is valid

---

**You can launch with this minimal setup and improve iteratively!** 🚀
