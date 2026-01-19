import Foundation
import Security

/// KeychainHelper provides secure storage for sensitive data like authentication tokens
/// Uses iOS Keychain instead of UserDefaults for encryption and security
class KeychainHelper {

    static let shared = KeychainHelper()

    private init() {}

    // MARK: - Save

    /// Saves a string value to the Keychain
    /// - Parameters:
    ///   - value: The string value to save
    ///   - key: The key to store the value under
    /// - Returns: True if successful, false otherwise
    @discardableResult
    func save(_ value: String, forKey key: String) -> Bool {
        guard let data = value.data(using: .utf8) else {
            print("🔴 KeychainHelper: Failed to encode value for key: \(key)")
            return false
        }

        // Delete any existing value first
        delete(forKey: key)

        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: key,
            kSecValueData as String: data,
            kSecAttrAccessible as String: kSecAttrAccessibleAfterFirstUnlock
        ]

        let status = SecItemAdd(query as CFDictionary, nil)

        if status == errSecSuccess {
            print("🟢 KeychainHelper: Successfully saved value for key: \(key)")
            return true
        } else {
            print("🔴 KeychainHelper: Failed to save value for key: \(key), status: \(status)")
            return false
        }
    }

    // MARK: - Retrieve

    /// Retrieves a string value from the Keychain
    /// - Parameter key: The key to retrieve
    /// - Returns: The string value if found, nil otherwise
    func retrieve(forKey key: String) -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)

        guard status == errSecSuccess,
              let data = result as? Data,
              let value = String(data: data, encoding: .utf8) else {
            if status != errSecItemNotFound {
                print("🔴 KeychainHelper: Failed to retrieve value for key: \(key), status: \(status)")
            }
            return nil
        }

        return value
    }

    // MARK: - Delete

    /// Deletes a value from the Keychain
    /// - Parameter key: The key to delete
    /// - Returns: True if successful or item doesn't exist, false on error
    @discardableResult
    func delete(forKey key: String) -> Bool {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: key
        ]

        let status = SecItemDelete(query as CFDictionary)

        if status == errSecSuccess || status == errSecItemNotFound {
            return true
        } else {
            print("🔴 KeychainHelper: Failed to delete value for key: \(key), status: \(status)")
            return false
        }
    }

    // MARK: - Clear All

    /// Clears all Keychain items stored by this app
    /// Use with caution - this will remove all stored credentials
    @discardableResult
    func clearAll() -> Bool {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword
        ]

        let status = SecItemDelete(query as CFDictionary)

        if status == errSecSuccess || status == errSecItemNotFound {
            print("🟢 KeychainHelper: Cleared all Keychain items")
            return true
        } else {
            print("🔴 KeychainHelper: Failed to clear Keychain, status: \(status)")
            return false
        }
    }
}

// MARK: - Keychain Keys

extension KeychainHelper {
    /// Standard keys for common credentials
    struct Keys {
        static let authToken = "com.gigco.authToken"
        static let userEmail = "com.gigco.userEmail"
        static let userId = "com.gigco.userId"
        static let userRole = "com.gigco.userRole"
    }
}
