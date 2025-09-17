//
//  JobService.swift
//  GigCo-Mobile
//
//  Created by Claude on 9/9/25.
//

import Foundation

@MainActor
class JobService: ObservableObject {
    @Published var jobs: [Job] = []
    @Published var availableJobs: [Job] = []
    @Published var myJobs: [Job] = []
    @Published var isLoading = false
    
    private let apiService = APIService.shared
    
    // MARK: - Job Listing Functions
    
    func getAllJobs() async throws {
        isLoading = true

        do {
            let response = try await apiService.getJobs()
            print("🟢 Jobs API response: \(response.jobs.count) jobs received")

            // Convert JobResponse to Job model
            self.jobs = response.jobs.map { jobResponse in
                Job(
                    id: jobResponse.id,
                    uuid: jobResponse.uuid,
                    title: jobResponse.title,
                    description: jobResponse.description,
                    category: jobResponse.category,
                    location: jobResponse.locationAddress,
                    price: jobResponse.totalPay,
                    status: jobResponse.status,
                    customerId: jobResponse.consumerID,
                    gigworkerId: nil, // Not included in list response
                    createdAt: jobResponse.createdAt,
                    updatedAt: jobResponse.updatedAt,
                    scheduledFor: jobResponse.scheduledStart
                )
            }

            print("🟢 Successfully converted \(self.jobs.count) jobs")
        } catch {
            print("🔴 Failed to fetch all jobs: \(error)")
            self.jobs = []
            throw error
        }

        isLoading = false
    }
    
    func getAvailableJobs() async throws {
        // Filter available jobs from all jobs based on status
        try await getAllJobs()
        self.availableJobs = jobs.filter { $0.status == "posted" || $0.status == "open" || $0.status == "available" }
    }
    
    func getMyJobs() async throws {
        // Note: This method will be called from views that have access to AuthService
        // For now, we'll return empty and let the view handle filtering
        self.myJobs = []
    }

    func getMyJobs(for userID: Int) async throws {
        // Get all jobs first, then filter by current user
        try await getAllJobs()

        // Filter jobs where current user is the consumer (posted by current user)
        self.myJobs = self.jobs.filter { job in
            job.customerId == userID
        }

        print("🟢 Found \(self.myJobs.count) jobs for user ID \(userID)")
    }
    
    func getJobById(_ id: Int) async throws -> Job? {
        // Since we don't have a specific job detail API in the generated client,
        // we'll find it from the loaded jobs list
        return jobs.first { $0.id == id }
    }
    
    // MARK: - Job Management (Simplified - API endpoints not available)
    
    func refreshJobs() async {
        do {
            try await getAllJobs()
        } catch {
            print("🔴 Failed to refresh jobs: \(error)")
        }
    }
    
    func createJob(_ jobRequest: CreateJobRequest, consumerID: Int) async throws -> JobCreateResponse {
        print("🔵 JobService.createJob - Creating job with title: \(jobRequest.title)")

        let scheduledStart = jobRequest.scheduledFor

        let response = try await apiService.createJob(
            title: jobRequest.title,
            description: jobRequest.description,
            category: jobRequest.category,
            location: jobRequest.location,
            price: jobRequest.price,
            consumerID: consumerID,
            scheduledStart: scheduledStart
        )

        print("🟢 JobService.createJob - Job created successfully: \(response)")

        // Add the new job to our local jobs list immediately
        let newJob = Job(
            id: response.id,
            uuid: response.uuid,
            title: response.title,
            description: response.description,
            category: response.category,
            location: response.locationAddress,
            price: response.totalPay,
            status: response.status,
            customerId: response.consumerID ?? consumerID,
            gigworkerId: nil,
            createdAt: response.createdAt,
            updatedAt: response.updatedAt,
            scheduledFor: response.scheduledStart
        )

        // Add to jobs list
        self.jobs.append(newJob)

        // If this is the current user's job, add to myJobs
        self.myJobs.append(newJob)

        return response
    }
    
    func acceptJob(_ jobId: Int) async throws {
        // TODO: Implement job acceptance when API endpoint is available
        print("🔵 Would accept job ID: \(jobId)")
        // For now, just simulate success
        try await Task.sleep(nanoseconds: 500_000_000) // 0.5 second delay
        print("🟢 Job acceptance simulated successfully")
    }
    
    func startJob(_ jobId: Int) async throws {
        // TODO: Implement job start when API endpoint is available
        print("🔵 Would start job ID: \(jobId)")
        try await Task.sleep(nanoseconds: 500_000_000) // 0.5 second delay
        print("🟢 Job start simulated successfully")
    }
    
    func completeJob(_ jobId: Int) async throws {
        // TODO: Implement job completion when API endpoint is available
        print("🔵 Would complete job ID: \(jobId)")
        try await Task.sleep(nanoseconds: 500_000_000) // 0.5 second delay
        print("🟢 Job completion simulated successfully")
    }
    
    func cancelJob(_ jobId: Int) async throws {
        // TODO: Implement job cancellation when API endpoint is available
        print("🔵 Would cancel job ID: \(jobId)")
        try await Task.sleep(nanoseconds: 500_000_000) // 0.5 second delay
        print("🟢 Job cancellation simulated successfully")
    }
}

// MARK: - Data Models

struct Job: Codable, Identifiable {
    let id: Int?
    let uuid: String?
    let title: String
    let description: String
    let category: String?
    let location: String?
    let price: Double?
    let status: String?
    let customerId: Int?
    let gigworkerId: Int?
    let createdAt: String?
    let updatedAt: String?
    let scheduledFor: String?
    
    enum CodingKeys: String, CodingKey {
        case id, uuid, title, description, category, location, price, status
        case customerId = "customer_id"
        case gigworkerId = "gigworker_id"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
        case scheduledFor = "scheduled_for"
    }
}

struct CreateJobRequest: Codable {
    let title: String
    let description: String
    let category: String
    let location: String
    let price: Double
    let scheduledFor: String?
    
    enum CodingKeys: String, CodingKey {
        case title, description, category, location, price
        case scheduledFor = "scheduled_for"
    }
}

