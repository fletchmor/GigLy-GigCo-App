import Foundation

/// Custom URLSession delegate to handle self-signed SSL certificates in development
/// WARNING: Only use this in development with DEBUG builds!
#if DEBUG
class SelfSignedCertDelegate: NSObject, URLSessionDelegate {

    func urlSession(_ session: URLSession,
                   didReceive challenge: URLAuthenticationChallenge,
                   completionHandler: @escaping (URLSession.AuthChallengeDisposition, URLCredential?) -> Void) {

        // Only bypass certificate validation for server trust challenges
        guard challenge.protectionSpace.authenticationMethod == NSURLAuthenticationMethodServerTrust else {
            completionHandler(.performDefaultHandling, nil)
            return
        }

        // Only bypass for development environment
        guard Configuration.Environment.current == .development else {
            completionHandler(.performDefaultHandling, nil)
            return
        }

        // Log the certificate details (helpful for debugging)
        if let serverTrust = challenge.protectionSpace.serverTrust {
            print("🔐 Accepting self-signed certificate for: \(challenge.protectionSpace.host)")

            // Create credential with the server trust
            let credential = URLCredential(trust: serverTrust)
            completionHandler(.useCredential, credential)
        } else {
            completionHandler(.performDefaultHandling, nil)
        }
    }
}
#endif

/// Production-ready URLSession delegate with SSL certificate pinning
///
/// To generate the public key hash for pinning:
/// 1. Get your server's certificate: openssl s_client -connect api.gigco.app:443 </dev/null 2>/dev/null | openssl x509 -outform DER > cert.der
/// 2. Extract public key: openssl x509 -inform DER -in cert.der -pubkey -noout > pubkey.pem
/// 3. Get SHA256 hash: openssl pkey -pubin -in pubkey.pem -outform DER | openssl dgst -sha256 -binary | base64
class SecureURLSessionDelegate: NSObject, URLSessionDelegate {

    // MARK: - Pinned Public Key Hashes
    // Add your production certificate's public key hash here
    // Generate with: openssl x509 -inform DER -in cert.der -pubkey -noout | openssl pkey -pubin -outform DER | openssl dgst -sha256 -binary | base64
    private static let pinnedPublicKeyHashes: [String] = [
        // Primary certificate hash - UPDATE THIS with your actual production cert hash
        "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", // Placeholder - replace with real hash
        // Backup certificate hash (for certificate rotation)
        "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=", // Placeholder - replace with real hash
    ]

    // Allowed hosts for pinning
    private static let pinnedHosts: Set<String> = [
        "api.gigco.app",
        "staging-api.gigco.app"
    ]

    func urlSession(_ session: URLSession,
                   didReceive challenge: URLAuthenticationChallenge,
                   completionHandler: @escaping (URLSession.AuthChallengeDisposition, URLCredential?) -> Void) {

        guard challenge.protectionSpace.authenticationMethod == NSURLAuthenticationMethodServerTrust,
              let serverTrust = challenge.protectionSpace.serverTrust else {
            completionHandler(.performDefaultHandling, nil)
            return
        }

        let host = challenge.protectionSpace.host

        // Only apply pinning to our known hosts
        guard Self.pinnedHosts.contains(host) else {
            // For other hosts, use default certificate validation
            completionHandler(.performDefaultHandling, nil)
            return
        }

        // Skip pinning if not enabled (e.g., in staging for testing)
        guard Configuration.isSSLPinningEnabled else {
            let credential = URLCredential(trust: serverTrust)
            completionHandler(.useCredential, credential)
            return
        }

        // Verify the certificate chain
        if validateCertificateChain(serverTrust: serverTrust, host: host) {
            let credential = URLCredential(trust: serverTrust)
            completionHandler(.useCredential, credential)
        } else {
            // Certificate validation failed - reject the connection
            print("⚠️ SSL Pinning failed for host: \(host)")
            completionHandler(.cancelAuthenticationChallenge, nil)
        }
    }

    /// Validates the certificate chain using public key pinning
    private func validateCertificateChain(serverTrust: SecTrust, host: String) -> Bool {
        // First, perform standard certificate validation
        var error: CFError?
        let isValid = SecTrustEvaluateWithError(serverTrust, &error)

        guard isValid else {
            print("⚠️ Standard certificate validation failed: \(error?.localizedDescription ?? "unknown")")
            return false
        }

        // Get the certificate chain
        guard let certificateChain = SecTrustCopyCertificateChain(serverTrust) as? [SecCertificate],
              !certificateChain.isEmpty else {
            print("⚠️ Could not get certificate chain")
            return false
        }

        // Check if any certificate in the chain matches our pinned hashes
        for certificate in certificateChain {
            if let publicKeyHash = getPublicKeyHash(from: certificate) {
                if Self.pinnedPublicKeyHashes.contains(publicKeyHash) {
                    return true
                }
            }
        }

        print("⚠️ No certificate matched pinned public keys")
        return false
    }

    /// Extracts and hashes the public key from a certificate
    private func getPublicKeyHash(from certificate: SecCertificate) -> String? {
        guard let publicKey = SecCertificateCopyKey(certificate) else {
            return nil
        }

        var error: Unmanaged<CFError>?
        guard let publicKeyData = SecKeyCopyExternalRepresentation(publicKey, &error) as Data? else {
            return nil
        }

        // Add ASN.1 header for RSA 2048 public key (if needed)
        // This depends on your certificate type
        let hashData = sha256(data: publicKeyData)
        return hashData.base64EncodedString()
    }

    /// Computes SHA256 hash of data
    private func sha256(data: Data) -> Data {
        var hash = [UInt8](repeating: 0, count: Int(CC_SHA256_DIGEST_LENGTH))
        data.withUnsafeBytes { buffer in
            _ = CC_SHA256(buffer.baseAddress, CC_LONG(data.count), &hash)
        }
        return Data(hash)
    }
}

// Import CommonCrypto for SHA256
import CommonCrypto
