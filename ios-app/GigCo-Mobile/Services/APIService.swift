//
//  APIService.swift
//  GigCo-Mobile
//
//  Created by Claude on 9/11/25.
//

import Foundation

@MainActor
class APIService: ObservableObject {
    static let shared = APIService()
    
    @Published var isConfigured = false
    private var authToken: String?
    
    private init() {
        configureAPI()
        loadAuthToken()
    }
    
    private func configureAPI() {
        // Configure the base URL for the API client
        // Use Mac's IP address for iOS simulator connectivity
        // Try 192.168.22.233 first, fallback to localhost for debugging
        GigCoAPIAPI.basePath = "http://192.168.22.233:8080/api/v1"
        
        // Configure authentication if token exists
        if let token = authToken {
            setAuthToken(token)
        }
        
        isConfigured = true
    }
    
    private func loadAuthToken() {
        authToken = UserDefaults.standard.string(forKey: "auth_token")
        if let token = authToken {
            setAuthToken(token)
        }
    }
    
    func setAuthToken(_ token: String) {
        authToken = token
        UserDefaults.standard.set(token, forKey: "auth_token")
        
        // Configure API client with Bearer token
        GigCoAPIAPI.customHeaders["Authorization"] = "Bearer \(token)"
    }
    
    func clearAuthToken() {
        authToken = nil
        UserDefaults.standard.removeObject(forKey: "auth_token")
        GigCoAPIAPI.customHeaders.removeValue(forKey: "Authorization")
    }
    
    // MARK: - Auth Methods
    
    func login(email: String, password: String) async throws -> ApiLoginResponse {
        print("🔵 APIService.login - Making direct URLSession request")
        print("🔵 APIService.login - Email: \(email)")
        
        // Create URL
        guard let url = URL(string: "http://192.168.22.233:8080/api/v1/auth/login") else {
            print("🔴 APIService.login - Invalid URL")
            throw APIError.invalidConfiguration
        }
        
        // Create request
        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = "POST"
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        // Create request body
        let requestBody = [
            "email": email,
            "password": password
        ]
        
        do {
            let jsonData = try JSONSerialization.data(withJSONObject: requestBody)
            urlRequest.httpBody = jsonData
            print("🔵 APIService.login - Request body created: \(String(data: jsonData, encoding: .utf8) ?? "nil")")
        } catch {
            print("🔴 APIService.login - Failed to serialize request body: \(error)")
            throw error
        }
        
        // Make the request
        return try await withCheckedThrowingContinuation { continuation in
            print("🔵 APIService.login - Starting URLSession request...")
            
            let task = URLSession.shared.dataTask(with: urlRequest) { data, response, error in
                print("🔵 APIService.login - URLSession completed")
                
                if let error = error {
                    print("🔴 APIService.login - Network error: \(error)")
                    continuation.resume(throwing: error)
                    return
                }
                
                guard let httpResponse = response as? HTTPURLResponse else {
                    print("🔴 APIService.login - Invalid response type")
                    continuation.resume(throwing: APIError.unexpectedResponse)
                    return
                }
                
                print("🔵 APIService.login - HTTP Status: \(httpResponse.statusCode)")
                
                guard let data = data else {
                    print("🔴 APIService.login - No response data")
                    continuation.resume(throwing: APIError.unexpectedResponse)
                    return
                }
                
                print("🔵 APIService.login - Response data: \(String(data: data, encoding: .utf8) ?? "nil")")
                
                // Parse response
                do {
                    let decoder = JSONDecoder()
                    decoder.keyDecodingStrategy = .convertFromSnakeCase
                    let loginResponse = try decoder.decode(ApiLoginResponse.self, from: data)
                    print("🟢 APIService.login - Successfully parsed response: \(loginResponse)")
                    continuation.resume(returning: loginResponse)
                } catch {
                    print("🔴 APIService.login - Failed to parse response: \(error)")
                    continuation.resume(throwing: error)
                }
            }
            
            task.resume()
            print("🔵 APIService.login - URLSession task started")
        }
    }
    
    func register(email: String, password: String, name: String, address: String, role: String) async throws -> ApiRegisterResponse {
        print("🔵 APIService.register - Making direct URLSession request")
        print("🔵 APIService.register - Email: \(email), Role: \(role)")

        // Create URL
        guard let url = URL(string: "http://192.168.22.233:8080/api/v1/auth/register") else {
            print("🔴 APIService.register - Invalid URL")
            throw APIError.invalidConfiguration
        }

        // Create request
        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = "POST"
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")

        // Create request body
        let requestBody = [
            "email": email,
            "password": password,
            "name": name,
            "address": address,
            "role": role
        ]

        do {
            let jsonData = try JSONSerialization.data(withJSONObject: requestBody)
            urlRequest.httpBody = jsonData
            print("🔵 APIService.register - Request body created: \(String(data: jsonData, encoding: .utf8) ?? "nil")")
        } catch {
            print("🔴 APIService.register - Failed to serialize request body: \(error)")
            throw error
        }

        // Make the request
        return try await withCheckedThrowingContinuation { continuation in
            print("🔵 APIService.register - Starting URLSession request...")

            let task = URLSession.shared.dataTask(with: urlRequest) { data, response, error in
                print("🔵 APIService.register - URLSession completed")

                if let error = error {
                    print("🔴 APIService.register - Network error: \(error)")
                    continuation.resume(throwing: error)
                    return
                }

                guard let httpResponse = response as? HTTPURLResponse else {
                    print("🔴 APIService.register - Invalid response type")
                    continuation.resume(throwing: APIError.unexpectedResponse)
                    return
                }

                print("🔵 APIService.register - HTTP Status: \(httpResponse.statusCode)")

                guard let data = data else {
                    print("🔴 APIService.register - No response data")
                    continuation.resume(throwing: APIError.unexpectedResponse)
                    return
                }

                print("🔵 APIService.register - Response data: \(String(data: data, encoding: .utf8) ?? "nil")")

                // Check for success status
                if httpResponse.statusCode >= 200 && httpResponse.statusCode < 300 {
                    // Parse response
                    do {
                        let decoder = JSONDecoder()
                        decoder.keyDecodingStrategy = .convertFromSnakeCase
                        let registerResponse = try decoder.decode(ApiRegisterResponse.self, from: data)
                        print("🟢 APIService.register - Successfully parsed response: \(registerResponse)")
                        continuation.resume(returning: registerResponse)
                    } catch {
                        print("🔴 APIService.register - Failed to parse response: \(error)")
                        continuation.resume(throwing: error)
                    }
                } else {
                    // Handle error response
                    let errorMessage = String(data: data, encoding: .utf8) ?? "Unknown error"
                    print("🔴 APIService.register - Server error: \(errorMessage)")
                    continuation.resume(throwing: APIError.serverError(httpResponse.statusCode, errorMessage))
                }
            }

            task.resume()
            print("🔵 APIService.register - URLSession task started")
        }
    }
    
    // MARK: - Jobs Methods

    func getJobs() async throws -> JobsListResponse {
        print("🔵 APIService.getJobs - Making direct URLSession request")

        // Create URL
        guard let url = URL(string: "http://192.168.22.233:8080/api/v1/jobs") else {
            print("🔴 APIService.getJobs - Invalid URL")
            throw APIError.invalidConfiguration
        }

        // Create request
        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = "GET"
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")

        // Add authorization header if token exists
        if let token = authToken {
            urlRequest.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        return try await withCheckedThrowingContinuation { continuation in
            let task = URLSession.shared.dataTask(with: urlRequest) { data, response, error in
                if let error = error {
                    print("🔴 APIService.getJobs - Network error: \(error)")
                    continuation.resume(throwing: error)
                    return
                }

                guard let data = data else {
                    print("🔴 APIService.getJobs - No data received")
                    continuation.resume(throwing: APIError.unexpectedResponse)
                    return
                }

                guard let httpResponse = response as? HTTPURLResponse else {
                    print("🔴 APIService.getJobs - Invalid HTTP response")
                    continuation.resume(throwing: APIError.unexpectedResponse)
                    return
                }

                print("🔵 APIService.getJobs - HTTP Status: \(httpResponse.statusCode)")

                // Check for success status
                if httpResponse.statusCode >= 200 && httpResponse.statusCode < 300 {
                    // Parse response
                    do {
                        let decoder = JSONDecoder()
                        let jobsResponse = try decoder.decode(JobsListResponse.self, from: data)
                        print("🟢 APIService.getJobs - Successfully parsed \(jobsResponse.jobs.count) jobs")
                        continuation.resume(returning: jobsResponse)
                    } catch {
                        print("🔴 APIService.getJobs - Failed to parse response: \(error)")
                        if let jsonString = String(data: data, encoding: .utf8) {
                            print("🔴 APIService.getJobs - Raw response JSON: \(jsonString)")
                        }
                        continuation.resume(throwing: error)
                    }
                } else {
                    // Handle error response
                    let errorMessage = String(data: data, encoding: .utf8) ?? "Unknown error"
                    print("🔴 APIService.getJobs - Server error: \(errorMessage)")
                    continuation.resume(throwing: APIError.serverError(httpResponse.statusCode, errorMessage))
                }
            }

            task.resume()
            print("🔵 APIService.getJobs - URLSession task started")
        }
    }

    // MARK: - User Methods
    
    func getUserProfile() async throws -> ModelUser {
        return try await withCheckedThrowingContinuation { continuation in
            UsersAPI.usersProfileGet { response, error in
                if let error = error {
                    continuation.resume(throwing: error)
                } else if let response = response {
                    continuation.resume(returning: response)
                } else {
                    continuation.resume(throwing: APIError.unexpectedResponse)
                }
            }
        }
    }
    
    // MARK: - Jobs Methods

    func getJobs(page: Int? = nil, limit: Int? = nil, status: String? = nil, location: String? = nil) async throws -> [String: AnyCodable] {
        return try await withCheckedThrowingContinuation { continuation in
            JobsAPI.jobsGet(page: page, limit: limit, status: status, location: location) { response, error in
                if let error = error {
                    continuation.resume(throwing: error)
                } else if let response = response {
                    continuation.resume(returning: response)
                } else {
                    continuation.resume(throwing: APIError.unexpectedResponse)
                }
            }
        }
    }

    func getMyJobs(userID: Int, role: String) async throws -> JobsListResponse {
        print("🔵 APIService.getMyJobs - UserID: \(userID), Role: \(role)")
        var components = URLComponents(string: "http://192.168.22.233:8080/api/v1/jobs/my-jobs")!
        components.queryItems = [
            URLQueryItem(name: "user_id", value: String(userID)),
            URLQueryItem(name: "role", value: role)
        ]

        guard let url = components.url else {
            throw APIError.invalidConfiguration
        }

        print("🔵 APIService.getMyJobs - URL: \(url.absoluteString)")
        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = "GET"
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if let token = authToken {
            print("🔵 APIService.getMyJobs - Using auth token: \(token.prefix(20))...")
            urlRequest.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        } else {
            print("🔴 APIService.getMyJobs - No auth token available")
        }

        return try await withCheckedThrowingContinuation { continuation in
            let task = URLSession.shared.dataTask(with: urlRequest) { data, response, error in
                if let error = error {
                    print("🔴 APIService.getMyJobs - Network error: \(error)")
                    continuation.resume(throwing: error)
                    return
                }

                guard let data = data else {
                    print("🔴 APIService.getMyJobs - No data received")
                    continuation.resume(throwing: APIError.unexpectedResponse)
                    return
                }

                guard let httpResponse = response as? HTTPURLResponse else {
                    print("🔴 APIService.getMyJobs - Invalid HTTP response")
                    continuation.resume(throwing: APIError.unexpectedResponse)
                    return
                }

                if httpResponse.statusCode >= 200 && httpResponse.statusCode < 300 {
                    do {
                        let response = try JSONDecoder().decode(JobsListResponse.self, from: data)
                        print("🟢 APIService.getMyJobs - Success: \(response.jobs.count) jobs")
                        continuation.resume(returning: response)
                    } catch {
                        print("🔴 APIService.getMyJobs - JSON decode error: \(error)")
                        print("🔴 APIService.getMyJobs - Raw data: \(String(data: data, encoding: .utf8) ?? "nil")")
                        continuation.resume(throwing: error)
                    }
                } else {
                    let errorBody = String(data: data, encoding: .utf8) ?? "No error body"
                    print("🔴 APIService.getMyJobs - HTTP error: \(httpResponse.statusCode)")
                    print("🔴 APIService.getMyJobs - Error body: \(errorBody)")
                    continuation.resume(throwing: APIError.serverError(httpResponse.statusCode, "Failed to get my jobs: \(errorBody)"))
                }
            }
            task.resume()
        }
    }

    func getAvailableJobs() async throws -> JobsListResponse {
        guard let url = URL(string: "http://192.168.22.233:8080/api/v1/jobs/available") else {
            throw APIError.invalidConfiguration
        }

        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = "GET"
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if let token = authToken {
            urlRequest.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        return try await withCheckedThrowingContinuation { continuation in
            let task = URLSession.shared.dataTask(with: urlRequest) { data, response, error in
                if let error = error {
                    print("🔴 APIService.getAvailableJobs - Network error: \(error)")
                    continuation.resume(throwing: error)
                    return
                }

                guard let data = data else {
                    print("🔴 APIService.getAvailableJobs - No data received")
                    continuation.resume(throwing: APIError.unexpectedResponse)
                    return
                }

                guard let httpResponse = response as? HTTPURLResponse else {
                    print("🔴 APIService.getAvailableJobs - Invalid HTTP response")
                    continuation.resume(throwing: APIError.unexpectedResponse)
                    return
                }

                if httpResponse.statusCode >= 200 && httpResponse.statusCode < 300 {
                    do {
                        let response = try JSONDecoder().decode(JobsListResponse.self, from: data)
                        print("🟢 APIService.getAvailableJobs - Success: \(response.jobs.count) jobs")
                        continuation.resume(returning: response)
                    } catch {
                        print("🔴 APIService.getAvailableJobs - JSON decode error: \(error)")
                        continuation.resume(throwing: error)
                    }
                } else {
                    print("🔴 APIService.getAvailableJobs - HTTP error: \(httpResponse.statusCode)")
                    continuation.resume(throwing: APIError.serverError(httpResponse.statusCode, "Failed to get available jobs"))
                }
            }
            task.resume()
        }
    }

    func acceptJob(jobID: Int, gigWorkerID: Int) async throws -> [String: Any] {
        guard let url = URL(string: "http://192.168.22.233:8080/api/v1/jobs/\(jobID)/accept") else {
            throw APIError.invalidConfiguration
        }

        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = "POST"
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if let token = authToken {
            urlRequest.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        let requestBody = ["gig_worker_id": gigWorkerID]
        urlRequest.httpBody = try JSONSerialization.data(withJSONObject: requestBody)

        return try await withCheckedThrowingContinuation { continuation in
            let task = URLSession.shared.dataTask(with: urlRequest) { data, response, error in
                if let error = error {
                    print("🔴 APIService.acceptJob - Network error: \(error)")
                    continuation.resume(throwing: error)
                    return
                }

                guard let data = data else {
                    print("🔴 APIService.acceptJob - No data received")
                    continuation.resume(throwing: APIError.unexpectedResponse)
                    return
                }

                guard let httpResponse = response as? HTTPURLResponse else {
                    print("🔴 APIService.acceptJob - Invalid HTTP response")
                    continuation.resume(throwing: APIError.unexpectedResponse)
                    return
                }

                if httpResponse.statusCode >= 200 && httpResponse.statusCode < 300 {
                    do {
                        let response = try JSONSerialization.jsonObject(with: data) as? [String: Any] ?? [:]
                        print("🟢 APIService.acceptJob - Success")
                        continuation.resume(returning: response)
                    } catch {
                        print("🔴 APIService.acceptJob - JSON decode error: \(error)")
                        continuation.resume(throwing: error)
                    }
                } else {
                    print("🔴 APIService.acceptJob - HTTP error: \(httpResponse.statusCode)")
                    continuation.resume(throwing: APIError.serverError(httpResponse.statusCode, "Failed to accept job"))
                }
            }
            task.resume()
        }
    }

    func createJob(title: String, description: String, category: String, location: String, price: Double, consumerID: Int, scheduledStart: String? = nil) async throws -> JobCreateResponse {
        print("🔵 APIService.createJob - Making direct URLSession request")
        print("🔵 APIService.createJob - Title: \(title), Category: \(category)")
        print("🔵 APIService.createJob - ConsumerID: \(consumerID)")

        // Create URL
        guard let url = URL(string: "http://192.168.22.233:8080/api/v1/jobs/create") else {
            print("🔴 APIService.createJob - Invalid URL")
            throw APIError.invalidConfiguration
        }

        // Create request
        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = "POST"
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")

        // Add authorization header if token exists
        if let token = authToken {
            urlRequest.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        // Create request body
        var requestBody: [String: Any] = [
            "title": title,
            "description": description,
            "category": category,
            "location_address": location,
            "total_pay": price,
            "consumer_id": consumerID
        ]

        if let scheduledStart = scheduledStart {
            requestBody["scheduled_start"] = scheduledStart
        }

        do {
            let jsonData = try JSONSerialization.data(withJSONObject: requestBody)
            urlRequest.httpBody = jsonData
            print("🔵 APIService.createJob - Request body created: \(String(data: jsonData, encoding: .utf8) ?? "nil")")
        } catch {
            print("🔴 APIService.createJob - Failed to serialize request body: \(error)")
            throw error
        }

        // Make the request
        return try await withCheckedThrowingContinuation { continuation in
            print("🔵 APIService.createJob - Starting URLSession request...")

            let task = URLSession.shared.dataTask(with: urlRequest) { data, response, error in
                print("🔵 APIService.createJob - URLSession completed")

                if let error = error {
                    print("🔴 APIService.createJob - Network error: \(error)")
                    continuation.resume(throwing: error)
                    return
                }

                guard let httpResponse = response as? HTTPURLResponse else {
                    print("🔴 APIService.createJob - Invalid response type")
                    continuation.resume(throwing: APIError.unexpectedResponse)
                    return
                }

                print("🔵 APIService.createJob - HTTP Status: \(httpResponse.statusCode)")

                guard let data = data else {
                    print("🔴 APIService.createJob - No response data")
                    continuation.resume(throwing: APIError.unexpectedResponse)
                    return
                }

                print("🔵 APIService.createJob - Response data: \(String(data: data, encoding: .utf8) ?? "nil")")

                // Check for success status
                if httpResponse.statusCode >= 200 && httpResponse.statusCode < 300 {
                    // Parse response
                    do {
                        let decoder = JSONDecoder()
                        // Don't use convertFromSnakeCase since we're manually mapping keys
                        let jobResponse = try decoder.decode(JobCreateResponse.self, from: data)
                        print("🟢 APIService.createJob - Successfully parsed response: \(jobResponse)")
                        print("🟢 APIService.createJob - Created job with consumer_id: \(jobResponse.consumerID ?? -1)")
                        continuation.resume(returning: jobResponse)
                    } catch {
                        print("🔴 APIService.createJob - Failed to parse response: \(error)")
                        if let jsonString = String(data: data, encoding: .utf8) {
                            print("🔴 APIService.createJob - Raw response JSON: \(jsonString)")
                        }
                        continuation.resume(throwing: error)
                    }
                } else {
                    // Handle error response
                    let errorMessage = String(data: data, encoding: .utf8) ?? "Unknown error"
                    print("🔴 APIService.createJob - Server error: \(errorMessage)")
                    continuation.resume(throwing: APIError.serverError(httpResponse.statusCode, errorMessage))
                }
            }

            task.resume()
            print("🔵 APIService.createJob - URLSession task started")
        }
    }

    // MARK: - GigWorker Profile Methods

    func createGigWorkerProfile(name: String, email: String, bio: String, hourlyRate: Double, experienceYears: Int, phone: String? = nil, address: String? = nil) async throws -> GigWorkerCreateResponse {
        print("🔵 APIService.createGigWorkerProfile - Making direct URLSession request")
        print("🔵 APIService.createGigWorkerProfile - Name: \(name), Email: \(email)")

        // Create URL
        guard let url = URL(string: "http://192.168.22.233:8080/api/v1/gigworkers/create") else {
            print("🔴 APIService.createGigWorkerProfile - Invalid URL")
            throw APIError.invalidConfiguration
        }

        // Create request
        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = "POST"
        urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")

        // Add authorization header if token exists
        if let token = authToken {
            urlRequest.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        // Create request body
        var requestBody: [String: Any] = [
            "name": name,
            "email": email,
            "bio": bio,
            "hourly_rate": hourlyRate,
            "experience_years": experienceYears
        ]

        if let phone = phone {
            requestBody["phone"] = phone
        }

        if let address = address {
            requestBody["address"] = address
        }

        do {
            let jsonData = try JSONSerialization.data(withJSONObject: requestBody)
            urlRequest.httpBody = jsonData
            print("🔵 APIService.createGigWorkerProfile - Request body created: \(String(data: jsonData, encoding: .utf8) ?? "nil")")
        } catch {
            print("🔴 APIService.createGigWorkerProfile - Failed to serialize request body: \(error)")
            throw error
        }

        // Make the request
        return try await withCheckedThrowingContinuation { continuation in
            print("🔵 APIService.createGigWorkerProfile - Starting URLSession request...")

            let task = URLSession.shared.dataTask(with: urlRequest) { data, response, error in
                print("🔵 APIService.createGigWorkerProfile - URLSession completed")

                if let error = error {
                    print("🔴 APIService.createGigWorkerProfile - Network error: \(error)")
                    continuation.resume(throwing: error)
                    return
                }

                guard let httpResponse = response as? HTTPURLResponse else {
                    print("🔴 APIService.createGigWorkerProfile - Invalid response type")
                    continuation.resume(throwing: APIError.unexpectedResponse)
                    return
                }

                print("🔵 APIService.createGigWorkerProfile - HTTP Status: \(httpResponse.statusCode)")

                guard let data = data else {
                    print("🔴 APIService.createGigWorkerProfile - No response data")
                    continuation.resume(throwing: APIError.unexpectedResponse)
                    return
                }

                print("🔵 APIService.createGigWorkerProfile - Response data: \(String(data: data, encoding: .utf8) ?? "nil")")

                // Check for success status
                if httpResponse.statusCode >= 200 && httpResponse.statusCode < 300 {
                    // Parse response
                    do {
                        let decoder = JSONDecoder()
                        decoder.keyDecodingStrategy = .convertFromSnakeCase
                        let gigWorkerResponse = try decoder.decode(GigWorkerCreateResponse.self, from: data)
                        print("🟢 APIService.createGigWorkerProfile - Successfully parsed response: \(gigWorkerResponse)")
                        continuation.resume(returning: gigWorkerResponse)
                    } catch {
                        print("🔴 APIService.createGigWorkerProfile - Failed to parse response: \(error)")
                        continuation.resume(throwing: error)
                    }
                } else {
                    // Handle error response
                    let errorMessage = String(data: data, encoding: .utf8) ?? "Unknown error"
                    print("🔴 APIService.createGigWorkerProfile - Server error: \(errorMessage)")
                    continuation.resume(throwing: APIError.serverError(httpResponse.statusCode, errorMessage))
                }
            }

            task.resume()
            print("🔵 APIService.createGigWorkerProfile - URLSession task started")
        }
    }
    
    // MARK: - Health Check
    
    func healthCheck() async throws -> [String: AnyCodable] {
        return try await withCheckedThrowingContinuation { continuation in
            HealthAPI.healthGet { response, error in
                if let error = error {
                    continuation.resume(throwing: error)
                } else if let response = response {
                    continuation.resume(returning: response)
                } else {
                    continuation.resume(throwing: APIError.unexpectedResponse)
                }
            }
        }
    }
}

enum APIError: LocalizedError {
    case unexpectedResponse
    case invalidConfiguration
    case serverError(Int, String)

    var errorDescription: String? {
        switch self {
        case .unexpectedResponse:
            return "Received unexpected response from server"
        case .invalidConfiguration:
            return "API client is not properly configured"
        case .serverError(let code, let message):
            return "Server error (\(code)): \(message)"
        }
    }
}

// MARK: - Response Models for API integration

struct JobCreateResponse: Codable {
    let id: Int
    let uuid: String
    let title: String
    let description: String
    let category: String?
    let locationAddress: String?
    let totalPay: Double?
    let status: String
    let consumerID: Int?
    let scheduledStart: String?
    let createdAt: String
    let updatedAt: String

    enum CodingKeys: String, CodingKey {
        case id, uuid, title, description, category, status
        case locationAddress = "location_address"
        case totalPay = "total_pay"
        case consumerID = "consumer_id"
        case scheduledStart = "scheduled_start"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }
}

struct GigWorkerCreateResponse: Codable {
    let id: Int
    let uuid: String
    let name: String
    let email: String
    let bio: String?
    let hourlyRate: Double?
    let experienceYears: Int?
    let verificationStatus: String
    let isActive: Bool
    let createdAt: String
    let updatedAt: String

    enum CodingKeys: String, CodingKey {
        case id, uuid, name, email, bio
        case hourlyRate = "hourly_rate"
        case experienceYears = "experience_years"
        case verificationStatus = "verification_status"
        case isActive = "is_active"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }
}

// MARK: - Jobs List Response Models

struct JobsListResponse: Codable {
    let jobs: [JobResponse]
    let pagination: Pagination

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)

        // Handle null jobs array by providing empty array as default
        if container.contains(.jobs) {
            if try container.decodeNil(forKey: .jobs) {
                // Field exists but is null
                jobs = []
            } else {
                // Field exists and has value
                jobs = try container.decode([JobResponse].self, forKey: .jobs)
            }
        } else {
            // Field doesn't exist
            jobs = []
        }

        pagination = try container.decode(Pagination.self, forKey: .pagination)
    }

    enum CodingKeys: String, CodingKey {
        case jobs, pagination
    }
}

struct JobResponse: Codable {
    let id: Int
    let uuid: String
    let consumerID: Int
    let title: String
    let description: String
    let category: String?
    let locationAddress: String?
    let totalPay: Double?
    let status: String
    let scheduledStart: String?
    let createdAt: String
    let updatedAt: String
    let consumer: ConsumerSummary

    enum CodingKeys: String, CodingKey {
        case id, uuid, title, description, category, status
        case consumerID = "consumer_id"
        case locationAddress = "location_address"
        case totalPay = "total_pay"
        case scheduledStart = "scheduled_start"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
        case consumer
    }
}

struct ConsumerSummary: Codable {
    let id: Int
    let uuid: String
    let name: String
}

struct Pagination: Codable {
    let page: Int
    let limit: Int
    let total: Int
    let pages: Int
    let hasNext: Bool
    let hasPrev: Bool

    enum CodingKeys: String, CodingKey {
        case page, limit, total, pages
        case hasNext = "has_next"
        case hasPrev = "has_prev"
    }
}