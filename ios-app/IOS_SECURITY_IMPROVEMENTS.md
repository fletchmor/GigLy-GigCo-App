# iOS App Security Improvements

## Overview

Critical security vulnerabilities have been fixed to make the GigCo iOS app production-ready. This document outlines what was changed and what you need to know.

---

## ✅ Security Issues Fixed

### 1. **Secure Token Storage with iOS Keychain** ✅

**Problem:** Auth tokens were stored in UserDefaults (unencrypted, accessible to attackers).

**Solution:** Implemented `KeychainHelper` that uses iOS Keychain for encrypted storage.

**Files Changed:**
- Created: `/Services/KeychainHelper.swift`
- Updated: `/Services/AuthService.swift` (lines 23-40, 118-126, 128-146)
- Updated: `/Services/APIService.swift` (lines 39-44, 46-52, 54-58)

**How It Works:**
```swift
// Saving token (encrypted in Keychain)
KeychainHelper.shared.save(token, forKey: KeychainHelper.Keys.authToken)

// Retrieving token
let token = KeychainHelper.shared.retrieve(forKey: KeychainHelper.Keys.authToken)

// Deleting token
KeychainHelper.shared.delete(forKey: KeychainHelper.Keys.authToken)
```

---

### 2. **Native SecureField for Passwords** ✅

**Problem:** Custom password masking implementation instead of Apple's secure native field.

**Solution:** Replaced with SwiftUI's `SecureField` while preserving show/hide functionality.

**Files Changed:**
- Updated: `/Views/LoginView.swift` (lines 43-60)
- Updated: `/Views/RegistrationView.swift` (lines 39-73)

**Before:**
```swift
ZStack {
    TextField("Password", text: $password)
        .foregroundColor(showPassword ? Color.primary : .clear)
    if !showPassword && !password.isEmpty {
        Text(String(repeating: "•", count: password.count))
    }
}
```

**After:**
```swift
if showPassword {
    TextField("Password", text: $password)
} else {
    SecureField("Password", text: $password)
}
```

---

### 3. **Environment-Based API Configuration** ✅

**Problem:** Hardcoded API endpoints scattered throughout codebase (`http://192.168.22.233:8080`).

**Solution:** Created centralized `Configuration` system with environment-specific settings.

**Files Changed:**
- Created: `/Config/Configuration.swift`
- Updated: `/Services/APIService.swift` (replaced all hardcoded URLs)

**Environments:**
- **Development** (DEBUG builds): Uses local development server
- **Staging**: Uses staging environment (when available)
- **Production** (RELEASE builds): Uses production servers

**How to Use:**
```swift
// Automatically uses correct URL based on build configuration
let url = URL(string: "\(Configuration.apiBaseURL)/auth/login")

// Environment-specific features
if Configuration.isLoggingEnabled {
    print("Debug log")
}
```

---

### 4. **HTTPS Configuration** ⚠️

**Status:** URLs configured for HTTPS, but requires backend setup.

**What's Changed:**
- All API URLs now use `https://` instead of `http://`
- Development: `https://192.168.22.233:8080/api/v1`
- Production: `https://api.gigco.app/api/v1`

**⚠️ IMPORTANT: Local Development Setup Required**

Since your backend is currently HTTP-only, you have **two options**:

#### **Option A: Temporarily Use HTTP for Development (Quick)**

1. Edit `/Config/Configuration.swift`
2. Change development URL from:
   ```swift
   return "https://192.168.22.233:8080/api/v1"
   ```
   To:
   ```swift
   return "http://192.168.22.233:8080/api/v1"
   ```
3. Do the same for `healthCheckURL`

**⚠️ Security Warning:** Only do this for local development. Production MUST use HTTPS.

#### **Option B: Set Up HTTPS for Local Development (Recommended)**

This is more work but closer to production environment.

**Steps:**

1. **Generate self-signed SSL certificate:**
   ```bash
   cd /Users/fletcher/app
   mkdir -p certs
   openssl req -x509 -newkey rsa:4096 -nodes \
     -keyout certs/key.pem \
     -out certs/cert.pem \
     -days 365 \
     -subj "/CN=192.168.22.233"
   ```

2. **Update docker-compose.yml** to use HTTPS:
   ```yaml
   services:
     app:
       ports:
         - "8443:8443"  # HTTPS port
       volumes:
         - ./certs:/app/certs
       environment:
         - PORT=8443
         - TLS_CERT=/app/certs/cert.pem
         - TLS_KEY=/app/certs/key.pem
   ```

3. **Update Go backend** to serve HTTPS:
   ```go
   // cmd/main.go
   func main() {
       r := chi.NewRouter()
       // ... your routes

       port := os.Getenv("PORT")
       tlsCert := os.Getenv("TLS_CERT")
       tlsKey := os.Getenv("TLS_KEY")

       if tlsCert != "" && tlsKey != "" {
           log.Printf("Starting HTTPS server on :%s", port)
           log.Fatal(http.ListenAndServeTLS(":"+port, tlsCert, tlsKey, r))
       } else {
           log.Printf("Starting HTTP server on :%s", port)
           log.Fatal(http.ListenAndServe(":"+port, r))
       }
   }
   ```

4. **Update iOS Configuration.swift**:
   ```swift
   case .development:
       return "https://192.168.22.233:8443/api/v1"  // Note: 8443 instead of 8080
   ```

5. **Disable SSL certificate validation for development** (self-signed certs):

   Add to `APIService.swift`:
   ```swift
   // Only for development with self-signed certificates
   #if DEBUG
   class SelfSignedCertDelegate: NSObject, URLSessionDelegate {
       func urlSession(_ session: URLSession,
                      didReceive challenge: URLAuthenticationChallenge,
                      completionHandler: @escaping (URLSession.AuthChallengeDisposition, URLCredential?) -> Void) {
           if challenge.protectionSpace.authenticationMethod == NSURLAuthenticationMethodServerTrust,
              let serverTrust = challenge.protectionSpace.serverTrust {
               completionHandler(.useCredential, URLCredential(trust: serverTrust))
           } else {
               completionHandler(.performDefaultHandling, nil)
           }
       }
   }
   #endif
   ```

---

## 📋 Configuration Reference

### Environment Configuration

Edit `/Config/Configuration.swift` to customize for your setup:

```swift
case .development:
    // Update this IP to match your development machine
    return "https://192.168.22.233:8080/api/v1"

case .staging:
    // Update when you have a staging server
    return "https://staging-api.gigco.app/api/v1"

case .production:
    // Update with your production domain
    return "https://api.gigco.app/api/v1"
```

### Build Configurations

- **Debug builds** (Xcode "Run"): Uses `.development` environment
- **Release builds** (Xcode "Archive"): Uses `.production` environment

### Feature Flags

Control app behavior by environment:

```swift
Configuration.isLoggingEnabled        // true in dev/staging, false in production
Configuration.showDetailedErrors      // true in dev/staging, false in production
Configuration.isSSLPinningEnabled     // false in dev, true in staging/production
Configuration.requestTimeout          // 60s in dev, 30s in production
```

---

## 🚀 Production Deployment Checklist

Before deploying to App Store:

- [ ] **Update production API URL** in `Configuration.swift`
- [ ] **Verify HTTPS is working** for production backend
- [ ] **Test with release build** (not just debug)
- [ ] **Remove all debug logging** or verify `isLoggingEnabled` is false in production
- [ ] **Enable SSL certificate pinning** for production
- [ ] **Test token storage** (login, logout, app restart)
- [ ] **Test SecureField** password masking
- [ ] **Verify no hardcoded credentials** in source code

---

## 🔐 Security Best Practices Implemented

1. ✅ **Encrypted storage** for sensitive data (Keychain)
2. ✅ **Native password fields** (SecureField)
3. ✅ **Environment-based configuration** (no hardcoded production values)
4. ✅ **HTTPS enforcement** for production
5. ✅ **Conditional logging** (disabled in production)
6. ⚠️ **SSL certificate pinning** (configured, needs production certificates)

---

## 📝 Testing

### Test Keychain Storage

1. Login to the app
2. Force quit the app
3. Relaunch the app
4. ✅ You should still be logged in (token retrieved from Keychain)

### Test SecureField

1. Go to login screen
2. Type password - should show bullets (••••)
3. Tap eye icon - should show plain text
4. Tap again - back to bullets

### Test Environment Configuration

1. Build and run in **Debug** mode
2. Check console for: `Environment: development`
3. Archive and export **Release** build
4. Check console for: `Environment: production`

---

## 🆘 Troubleshooting

### "SSL certificate problem: self signed certificate"

**Cause:** Using HTTPS with self-signed certificate.

**Solution:** Either:
- Use HTTP for development (Option A above)
- Disable SSL validation for development (Option B above)

### "Failed to connect to server"

**Cause:** Backend is running on HTTP but app expects HTTPS (or vice versa).

**Solution:**
1. Check backend logs: `docker compose logs app`
2. Verify port: `docker compose ps`
3. Match URLs in `Configuration.swift` with actual backend

### "Token not persisting between app launches"

**Cause:** Keychain access may be blocked (simulator restrictions).

**Solution:**
- Reset simulator: Device → Erase All Content and Settings
- Test on physical device for more accurate Keychain behavior

### "Cannot build - Configuration not found"

**Cause:** New file not added to Xcode project.

**Solution:**
1. In Xcode, right-click on project navigator
2. Add Files to "GigCo-Mobile"
3. Select `Config/Configuration.swift`
4. Ensure "Add to targets: GigCo-Mobile" is checked

---

## 📊 Impact Summary

| Security Issue | Severity | Status |
|---------------|----------|--------|
| Unencrypted token storage | 🔴 Critical | ✅ Fixed |
| Custom password field | 🟡 High | ✅ Fixed |
| Hardcoded API URLs | 🟡 High | ✅ Fixed |
| HTTP (no HTTPS) | 🔴 Critical | ⚠️ Needs backend setup |
| No environment separation | 🟡 High | ✅ Fixed |

---

## 🎯 Next Steps

1. **Immediate:** Choose HTTP (Option A) or HTTPS (Option B) for local development
2. **Before TestFlight:** Set up staging environment with HTTPS
3. **Before App Store:** Configure production environment with proper SSL certificates

---

## 📚 Additional Resources

- [Apple Keychain Documentation](https://developer.apple.com/documentation/security/keychain_services)
- [SecureField Documentation](https://developer.apple.com/documentation/swiftui/securefield)
- [App Transport Security](https://developer.apple.com/documentation/bundleresources/information_property_list/nsapptransportsecurity)
- [SSL Certificate Pinning Guide](https://www.raywenderlich.com/1484288-preventing-man-in-the-middle-attacks-in-ios-with-ssl-pinning)

---

**Last Updated:** 2025-12-15
**iOS Version:** 15.0+
**Xcode Version:** 14.0+
