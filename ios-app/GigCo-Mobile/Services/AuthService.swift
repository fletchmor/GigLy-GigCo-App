//
//  AuthService.swift
//  GigCo-Mobile
//
//  Created by Fletcher Morris on 9/8/25.
//

import Foundation

@MainActor
class AuthService: ObservableObject {
    @Published var isAuthenticated = false
    @Published var currentUser: Person?
    @Published var authToken: String?
    
    private let apiService = APIService.shared
    
    init() {
        // Register this instance with APIService for token expiration handling
        apiService.setAuthServiceReference(self)

        // Check for existing auth token and user data
        if let token = UserDefaults.standard.string(forKey: "auth_token") {
            self.authToken = token

            // Try to restore user data
            if let userData = UserDefaults.standard.data(forKey: "current_user"),
               let user = try? JSONDecoder().decode(Person.self, from: userData) {
                self.currentUser = user
                self.isAuthenticated = true
                print("🟢 AuthService.init - Restored user session: \(user.name), ID: \(user.id ?? -1)")
            } else {
                // Token exists but no user data, clear everything
                print("🟡 AuthService.init - Token exists but no user data, clearing session")
                UserDefaults.standard.removeObject(forKey: "auth_token")
                self.authToken = nil
                self.isAuthenticated = false
            }
        }
    }
    
    func register(email: String, password: String, firstName: String, lastName: String, role: PersonRole = .consumer, address: String = "") async throws {
        print("🔵 Starting registration for email: \(email), role: \(role.rawValue)")

        let finalAddress = address.isEmpty ? "Default Address" : address

        do {
            let response = try await apiService.register(
                email: email,
                password: password,
                name: "\(firstName) \(lastName)",
                address: finalAddress,
                role: role.rawValue
            )
            
            print("🟢 Registration successful: \(response)")
            
            if let token = response.token {
                print("🟢 Got token, creating user and setting auth state")
                print("🔵 Registration response - ID: \(response.id ?? -1), UUID: \(response.uuid ?? "nil"), Name: \(response.name ?? "nil")")
                let user = Person(
                    id: response.id,
                    uuid: response.uuid,
                    name: response.name ?? "\(firstName) \(lastName)",
                    email: response.email ?? email,
                    role: response.role ?? PersonRole.consumer.rawValue,
                    isActive: response.isActive,
                    emailVerified: response.emailVerified,
                    phoneVerified: response.phoneVerified
                )
                print("🔵 Created Person object - ID: \(user.id ?? -1), Name: \(user.name), Role: \(user.role)")
                await setAuthenticationState(token: token, user: user)
                print("🟢 Registration completed successfully")
            } else {
                print("🔴 No token in response")
                throw AuthError.decodingError
            }
        } catch {
            print("🔴 Registration error: \(error)")
            throw error
        }
    }
    
    func login(email: String, password: String) async throws {
        print("🔵 Starting login for email: \(email)")
        
        do {
            let response = try await apiService.login(email: email, password: password)
            print("🟢 Login successful: \(response)")
            
            if let token = response.token {
                print("🟢 Got token, creating user and setting auth state")
                print("🔵 Login response - ID: \(response.id ?? -1), UUID: \(response.uuid ?? "nil"), Name: \(response.name ?? "nil")")
                let user = Person(
                    id: response.id,
                    uuid: response.uuid,
                    name: response.name ?? "Unknown",
                    email: response.email ?? email,
                    role: response.role ?? PersonRole.consumer.rawValue,
                    isActive: response.isActive,
                    emailVerified: response.emailVerified,
                    phoneVerified: response.phoneVerified
                )
                print("🔵 Created Person object - ID: \(user.id ?? -1), Name: \(user.name), Role: \(user.role)")
                await setAuthenticationState(token: token, user: user)
                print("🟢 Login completed successfully")
            } else {
                print("🔴 No token in response")
                throw AuthError.decodingError
            }
        } catch {
            print("🔴 Login error: \(error)")
            throw error
        }
    }
    
    func logout() {
        apiService.clearAuthToken()
        UserDefaults.standard.removeObject(forKey: "auth_token")
        UserDefaults.standard.removeObject(forKey: "current_user")
        authToken = nil
        currentUser = nil
        isAuthenticated = false
        print("🟢 User logged out and session data cleared")
    }
    
    private func setAuthenticationState(token: String, user: Person) async {
        print("🔵 Setting authentication state - User ID: \(user.id ?? -1), Name: \(user.name)")
        apiService.setAuthToken(token)
        UserDefaults.standard.set(token, forKey: "auth_token")

        // Save user data
        if let userData = try? JSONEncoder().encode(user) {
            UserDefaults.standard.set(userData, forKey: "current_user")
            print("🟢 User data saved to UserDefaults")
        } else {
            print("🔴 Failed to encode user data")
        }

        self.authToken = token
        self.currentUser = user
        self.isAuthenticated = true
        print("🟢 Authentication state set - isAuthenticated: \(self.isAuthenticated), currentUser ID: \(self.currentUser?.id ?? -1)")
    }
}


enum PersonRole: String, Codable {
    case consumer = "consumer"
    case gigWorker = "gig_worker"
    case admin = "admin"
}

struct Person: Codable, Identifiable {
    let id: Int?
    let uuid: String?
    let name: String
    let email: String
    let role: String
    let isActive: Bool?
    let emailVerified: Bool?
    let phoneVerified: Bool?
    
    enum CodingKeys: String, CodingKey {
        case id, uuid, name, email, role
        case isActive = "is_active"
        case emailVerified = "email_verified"
        case phoneVerified = "phone_verified"
    }
}

enum AuthError: LocalizedError {
    case invalidResponse
    case serverError(Int)
    case serverErrorWithMessage(Int, String)
    case decodingError
    
    var errorDescription: String? {
        switch self {
        case .invalidResponse:
            return "Invalid server response"
        case .serverError(let code):
            return "Server error with code: \(code)"
        case .serverErrorWithMessage(let code, let message):
            return "Server error (\(code)): \(message)"
        case .decodingError:
            return "Failed to decode response"
        }
    }
}