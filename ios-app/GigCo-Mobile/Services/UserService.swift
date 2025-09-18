//
//  UserService.swift
//  GigCo-Mobile
//
//  Created by Claude on 9/11/25.
//

import Foundation

@MainActor
class UserService: ObservableObject {
    static let shared = UserService()
    
    @Published var userProfile: ModelUser?
    @Published var isLoading = false
    @Published var errorMessage: String?
    
    private let apiService = APIService.shared
    
    private init() {}
    
    func fetchUserProfile() async throws {
        isLoading = true
        errorMessage = nil
        
        do {
            let profile = try await apiService.getUserProfile()
            userProfile = profile
            print("🟢 User profile fetched successfully: \(profile)")
        } catch {
            errorMessage = error.localizedDescription
            print("🔴 Failed to fetch user profile: \(error)")
            throw error
        }
        
        isLoading = false
    }
    
    func refreshUserProfile() async {
        do {
            try await fetchUserProfile()
        } catch {
            print("🔴 Failed to refresh user profile: \(error)")
        }
    }
    
    func clearUserData() {
        userProfile = nil
        errorMessage = nil
    }

    func createGigWorkerProfile(name: String, email: String, bio: String, hourlyRate: Double, experienceYears: Int, phone: String? = nil, address: String? = nil) async throws -> GigWorkerCreateResponse {
        isLoading = true
        errorMessage = nil

        do {
            let response = try await apiService.createGigWorkerProfile(
                name: name,
                email: email,
                bio: bio,
                hourlyRate: hourlyRate,
                experienceYears: experienceYears,
                phone: phone,
                address: address
            )
            print("🟢 Gig worker profile created successfully: \(response)")
            isLoading = false
            return response
        } catch {
            errorMessage = error.localizedDescription
            print("🔴 Failed to create gig worker profile: \(error)")
            isLoading = false
            throw error
        }
    }
}