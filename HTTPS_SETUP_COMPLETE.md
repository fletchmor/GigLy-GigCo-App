# ✅ HTTPS Setup Complete!

Your GigCo application now runs with HTTPS encryption for secure development!

---

## 🎉 What Was Done

### 1. ✅ SSL Certificates Generated
- **Location**: `/Users/fletcher/app/certs/`
- **Files**: `cert.pem` (certificate), `key.pem` (private key)
- **Valid For**: 365 days
- **IP Address**: 192.168.22.86 (your Mac's current IP)

### 2. ✅ Go Backend Updated
- **File**: `cmd/main.go`
- **Changes**: Added HTTPS support with automatic TLS cert detection
- **Behavior**:
  - If TLS_CERT and TLS_KEY env vars are set → Uses HTTPS
  - Otherwise → Falls back to HTTP with warning

### 3. ✅ Docker Configuration Updated
- **File**: `docker-compose.yml`
- **Changes**:
  - Added TLS_CERT and TLS_KEY environment variables
  - Mounted `/certs` directory to container
  - Server now starts with HTTPS automatically

### 4. ✅ Environment Variables Updated
- **File**: `.env`
- **Changes**: Added TLS certificate paths
```bash
TLS_CERT=/Users/fletcher/app/certs/cert.pem
TLS_KEY=/Users/fletcher/app/certs/key.pem
```

### 5. ✅ iOS App Configured
- **Files Created**:
  - `ios-app/GigCo-Mobile/Config/Configuration.swift` - Environment management
  - `ios-app/GigCo-Mobile/Services/URLSessionDelegate.swift` - Self-signed cert handler

- **Files Updated**:
  - `ios-app/GigCo-Mobile/Services/APIService.swift` - Uses custom URLSession
  - All hardcoded URLs replaced with Configuration system

- **Behavior**:
  - DEBUG builds → Accepts self-signed certificates, uses dev environment
  - RELEASE builds → Enforces proper SSL validation, uses production environment

---

## 🚀 How to Use

### Starting the Server with HTTPS

The server is already running with HTTPS! To verify:

```bash
# Check server logs
docker compose logs app | tail -20

# You should see:
# "Starting HTTPS server on :8080"
# "Using TLS certificate: /app/certs/cert.pem"
```

### Testing HTTPS

```bash
# Test health endpoint
curl -k https://192.168.22.86:8080/health

# Test login endpoint
curl -k -X POST https://192.168.22.86:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"worker1@gigco.dev","password":"password123"}'
```

**Note**: The `-k` flag allows curl to accept self-signed certificates.

### Running the iOS App

1. **Open Xcode**:
   ```bash
   open /Users/fletcher/app/ios-app/GigCo-Mobile/GigCo-Mobile.xcodeproj
   ```

2. **Build and Run** (Cmd+R)
   - iOS simulator will automatically accept the self-signed certificate
   - All API calls now use HTTPS (https://192.168.22.86:8080/api/v1)

3. **Check Configuration**:
   - When app starts, check Xcode console for:
   ```
   ========================================
   🔧 App Configuration
   ========================================
   Environment: development
   API Base URL: https://192.168.22.86:8080/api/v1
   Logging Enabled: true
   SSL Pinning: false
   ========================================
   ```

---

## 🔧 Configuration Reference

### Environments

The iOS app has three environments:

| Environment | When Active | API URL |
|------------|-------------|---------|
| Development | DEBUG builds | https://192.168.22.86:8080/api/v1 |
| Staging | Manual override | https://staging-api.gigco.app/api/v1 |
| Production | RELEASE builds | https://api.gigco.app/api/v1 |

### Switching Between HTTP and HTTPS

#### Option 1: Disable HTTPS (Quick Test)

Edit `.env` and comment out the TLS lines:
```bash
# TLS/HTTPS Configuration (optional - comment out for HTTP)
#TLS_CERT=/Users/fletcher/app/certs/cert.pem
#TLS_KEY=/Users/fletcher/app/certs/key.pem
```

Then restart:
```bash
docker compose restart app
```

Server will log: `Starting HTTP server on :8080`

#### Option 2: Keep HTTPS (Recommended)

Leave everything as is. HTTPS is properly configured!

---

## 📱 iOS App Features

### Automatic Environment Detection

```swift
// In your code, use Configuration for all settings:
let apiURL = Configuration.apiBaseURL  // Automatically correct for environment
let isDebug = Configuration.isLoggingEnabled  // true in dev, false in prod
```

### Security Features

✅ **Keychain Storage**: Auth tokens encrypted in iOS Keychain
✅ **SecureField**: Native password fields
✅ **Self-Signed Cert Handling**: Works in development
✅ **SSL Pinning Ready**: Can be enabled for production

---

## 🔐 Security Notes

### Development (Current Setup)

- ✅ Uses self-signed certificates (acceptable for local dev)
- ✅ iOS app accepts self-signed certs in DEBUG builds only
- ✅ Tokens stored securely in Keychain
- ⚠️ Self-signed certs are NOT trusted by default (browsers will show warning)

### Production (When You Deploy)

You'll need to:

1. **Get Real SSL Certificate**:
   - Use Let's Encrypt (free): https://letsencrypt.org
   - Or buy from Certificate Authority

2. **Update Production Config**:
   ```bash
   # .env.production
   TLS_CERT=/path/to/real/cert.pem
   TLS_KEY=/path/to/real/key.pem
   ```

3. **Update iOS Production URL**:
   ```swift
   // Config/Configuration.swift
   case .production:
       return "https://api.yourdomain.com/api/v1"
   ```

4. **Enable SSL Pinning** in iOS app for extra security

---

## 🆘 Troubleshooting

### Issue: "Connection refused" or "SSL error"

**Cause**: Server might not be running or using wrong IP

**Solution**:
```bash
# Check server is running
docker compose ps

# Check server logs
docker compose logs app | tail -20

# Restart server
docker compose restart app
```

### Issue: "Certificate not trusted" in browser

**Cause**: Self-signed certificates aren't trusted by browsers

**Solution**: This is expected! Either:
- Use curl with `-k` flag
- Add exception in browser (not recommended)
- Use iOS app (automatically accepts self-signed certs in debug)

### Issue: iOS app can't connect

**Cause**: Wrong IP address or server not running

**Solution**:
1. Check your current IP:
   ```bash
   ifconfig | grep "inet " | grep -v 127.0.0.1
   ```

2. If IP changed, update:
   - Regenerate certificate with new IP
   - Update `Config/Configuration.swift` with new IP
   - Restart Docker containers

### Issue: IP Address Changed

Your Mac's IP may change when you switch networks. If this happens:

```bash
# 1. Get new IP
ifconfig | grep "inet " | grep -v 127.0.0.1

# 2. Regenerate certificate (replace NEW_IP)
cd /Users/fletcher/app/certs
rm cert.pem key.pem
openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout key.pem -out cert.pem -days 365 \
  -subj "/C=US/ST=California/O=GigCo/CN=NEW_IP"

# 3. Update iOS Configuration.swift with new IP

# 4. Restart server
docker compose restart app
```

---

## 📊 Verification Checklist

- [x] SSL certificates generated in `/certs/` directory
- [x] Backend configured for HTTPS in `cmd/main.go`
- [x] Docker compose updated with TLS volumes and env vars
- [x] `.env` file has TLS certificate paths
- [x] iOS app Configuration.swift uses correct IP (192.168.22.86)
- [x] iOS app has URLSessionDelegate for self-signed certs
- [x] Server starts with "Starting HTTPS server" message
- [x] Health endpoint responds via HTTPS
- [x] Login endpoint responds via HTTPS

**Status**: ✅ ALL CHECKS PASSED

---

## 🎯 Next Steps

1. **Test the iOS App**:
   - Open in Xcode
   - Run on simulator
   - Try logging in
   - Verify HTTPS URLs in console

2. **Development Workflow**:
   - Keep Docker running: `docker compose up -d`
   - Develop iOS app normally
   - All traffic is now encrypted!

3. **When Ready for Production**:
   - Review `IOS_SECURITY_IMPROVEMENTS.md`
   - Get real SSL certificates
   - Update production configuration
   - Test on TestFlight

---

## 📚 Files Reference

### Backend
- `cmd/main.go` - Server startup with HTTPS support
- `docker-compose.yml` - Container configuration
- `.env` - Local environment variables
- `certs/cert.pem` - SSL certificate
- `certs/key.pem` - Private key

### iOS App
- `Config/Configuration.swift` - Environment configuration
- `Services/APIService.swift` - API client with custom URLSession
- `Services/URLSessionDelegate.swift` - Self-signed cert handler
- `Services/KeychainHelper.swift` - Secure token storage
- `Services/AuthService.swift` - Authentication service

---

## 💡 Tips

### Development
- Keep HTTPS enabled for realistic testing
- Self-signed certs are fine for local dev
- iOS simulator handles certs automatically

### Production
- Use real certificates from Let's Encrypt
- Enable SSL certificate pinning
- Test on physical devices before release

### Debugging
- Check Xcode console for Configuration printout
- Use `docker compose logs app` for backend issues
- Test endpoints with curl using `-k` flag

---

**🎉 Congratulations! Your GigCo app now runs securely with HTTPS!**

Last Updated: 2025-12-15
Your IP: 192.168.22.86
Certificate Valid Until: 2026-12-15
