import Foundation

/// App configuration for different environments
///
/// For local development, you can override the API URL by setting:
/// - UserDefaults key: "DEV_API_HOST" (e.g., "192.168.1.100:8080")
/// - Or in Xcode: Edit Scheme > Run > Arguments > Environment Variables > DEV_API_HOST
enum Configuration {

    // MARK: - Environment Types

    enum Environment {
        case development
        case staging
        case production

        /// Current active environment
        static var current: Environment {
            #if DEBUG
            return .development
            #else
            return .production
            #endif
        }
    }

    // MARK: - Development URL Override

    /// Returns the development API host, checking for overrides
    private static var developmentHost: String {
        // Check UserDefaults first (set via Settings bundle or programmatically)
        if let host = UserDefaults.standard.string(forKey: "DEV_API_HOST"), !host.isEmpty {
            return host
        }

        // Check environment variable (set via Xcode scheme)
        if let host = ProcessInfo.processInfo.environment["DEV_API_HOST"], !host.isEmpty {
            return host
        }

        // Default fallback - update this to your Mac's current IP
        // Find your IP: System Settings > Wi-Fi > Details > IP Address
        return "localhost:8080"
    }

    // MARK: - API Configuration

    /// Base API URL for the current environment
    static var apiBaseURL: String {
        switch Environment.current {
        case .development:
            // For local development - configurable via DEV_API_HOST
            return "https://\(developmentHost)/api/v1"

        case .staging:
            // Staging environment
            return "https://staging-api.gigco.app/api/v1"

        case .production:
            // Production environment - UPDATE THIS before release
            return "https://api.gigco.app/api/v1"
        }
    }

    /// Health check endpoint URL
    static var healthCheckURL: String {
        switch Environment.current {
        case .development:
            return "https://\(developmentHost)/health"
        case .staging:
            return "https://staging-api.gigco.app/health"
        case .production:
            return "https://api.gigco.app/health"
        }
    }

    // MARK: - Feature Flags

    /// Whether to enable verbose logging
    static var isLoggingEnabled: Bool {
        switch Environment.current {
        case .development, .staging:
            return true
        case .production:
            return false
        }
    }

    /// Whether to show detailed error messages
    static var showDetailedErrors: Bool {
        switch Environment.current {
        case .development, .staging:
            return true
        case .production:
            return false
        }
    }

    // MARK: - App Settings

    /// App version
    static var appVersion: String {
        Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "1.0.0"
    }

    /// Build number
    static var buildNumber: String {
        Bundle.main.infoDictionary?["CFBundleVersion"] as? String ?? "1"
    }

    /// App display name
    static var appName: String {
        Bundle.main.infoDictionary?["CFBundleDisplayName"] as? String ?? "GigCo"
    }

    // MARK: - Network Settings

    /// Request timeout interval in seconds
    static var requestTimeout: TimeInterval {
        switch Environment.current {
        case .development:
            return 60.0 // Longer timeout for debugging
        case .staging, .production:
            return 30.0
        }
    }

    /// Maximum number of retry attempts for failed requests
    static var maxRetryAttempts: Int {
        return 3
    }

    // MARK: - Security Settings

    /// Whether SSL certificate pinning is enabled
    static var isSSLPinningEnabled: Bool {
        switch Environment.current {
        case .development:
            return false // Disabled for local development
        case .staging, .production:
            return true
        }
    }

    // MARK: - Utility Methods

    /// Prints configuration info (for debugging)
    static func printConfiguration() {
        print("========================================")
        print("🔧 App Configuration")
        print("========================================")
        print("Environment: \(Environment.current)")
        print("App Name: \(appName)")
        print("Version: \(appVersion) (\(buildNumber))")
        print("API Base URL: \(apiBaseURL)")
        print("Logging Enabled: \(isLoggingEnabled)")
        print("SSL Pinning: \(isSSLPinningEnabled)")
        print("========================================")
    }
}
