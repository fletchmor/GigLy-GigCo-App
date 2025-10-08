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
                    gigworkerId: jobResponse.gigWorkerID,
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
        isLoading = true

        do {
            let response = try await apiService.getAvailableJobs()
            print("🟢 Available Jobs API response: \(response.jobs.count) jobs received")

            // Convert JobResponse to Job model
            self.availableJobs = response.jobs.map { jobResponse in
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
                    gigworkerId: jobResponse.gigWorkerID,
                    createdAt: jobResponse.createdAt,
                    updatedAt: jobResponse.updatedAt,
                    scheduledFor: jobResponse.scheduledStart
                )
            }

            print("🟢 Successfully converted \(self.availableJobs.count) available jobs")
        } catch {
            print("🔴 Failed to fetch available jobs: \(error)")
            self.availableJobs = []
            throw error
        }

        isLoading = false
    }
    
    func getMyJobs() async throws {
        // Note: This method will be called from views that have access to AuthService
        // For now, we'll return empty and let the view handle filtering
        self.myJobs = []
    }

    func getMyJobs(for userID: Int, role: String) async throws {
        isLoading = true
        print("🔵 JobService.getMyJobs - Called with userID: \(userID), role: \(role)")

        do {
            let response = try await apiService.getMyJobs(userID: userID, role: role)
            print("🟢 My Jobs API response: \(response.jobs.count) jobs received")

            // Convert JobResponse to Job model
            self.myJobs = response.jobs.map { jobResponse in
                let job = Job(
                    id: jobResponse.id,
                    uuid: jobResponse.uuid,
                    title: jobResponse.title,
                    description: jobResponse.description,
                    category: jobResponse.category,
                    location: jobResponse.locationAddress,
                    price: jobResponse.totalPay,
                    status: jobResponse.status,
                    customerId: jobResponse.consumerID,
                    gigworkerId: jobResponse.gigWorkerID,
                    createdAt: jobResponse.createdAt,
                    updatedAt: jobResponse.updatedAt,
                    scheduledFor: jobResponse.scheduledStart
                )
                print("🔵 JobService - Mapped job: \(job.title), ID: \(job.id ?? -1), Creator: \(job.customerId ?? -1), Status: \(job.status ?? "nil")")
                return job
            }

            print("🟢 Successfully converted \(self.myJobs.count) my jobs")
        } catch {
            print("🔴 Failed to fetch my jobs: \(error)")
            self.myJobs = []
            throw error
        }

        isLoading = false
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

        print("🟢 JobService.createJob - Created job locally with customerId: \(newJob.customerId ?? -1)")
        print("🟢 JobService.createJob - Current user should be: \(consumerID)")

        // Add to jobs list
        self.jobs.append(newJob)

        // If this is the current user's job, add to myJobs
        self.myJobs.append(newJob)
        print("🟢 JobService.createJob - Added to myJobs. Total myJobs: \(self.myJobs.count)")

        // Notify dashboard to refresh
        NotificationCenter.default.post(name: NSNotification.Name("RefreshDashboard"), object: nil)

        return response
    }
    
    func acceptJob(_ jobId: Int, gigWorkerID: Int) async throws {
        print("🔵 Accepting job ID: \(jobId) for worker: \(gigWorkerID)")

        do {
            let response = try await apiService.acceptJob(jobID: jobId, gigWorkerID: gigWorkerID)
            print("🟢 Job acceptance successful: \(response)")

            // Update local job lists to reflect the acceptance
            if let jobIndex = availableJobs.firstIndex(where: { $0.id == jobId }) {
                var updatedJob = availableJobs[jobIndex]
                updatedJob = Job(
                    id: updatedJob.id,
                    uuid: updatedJob.uuid,
                    title: updatedJob.title,
                    description: updatedJob.description,
                    category: updatedJob.category,
                    location: updatedJob.location,
                    price: updatedJob.price,
                    status: "accepted",
                    customerId: updatedJob.customerId,
                    gigworkerId: gigWorkerID,
                    createdAt: updatedJob.createdAt,
                    updatedAt: updatedJob.updatedAt,
                    scheduledFor: updatedJob.scheduledFor
                )

                // Remove from available jobs and add to my jobs
                availableJobs.remove(at: jobIndex)
                myJobs.append(updatedJob)

                // Also update in main jobs list
                if let mainJobIndex = jobs.firstIndex(where: { $0.id == jobId }) {
                    jobs[mainJobIndex] = updatedJob
                }

                // Notify dashboard to refresh
                NotificationCenter.default.post(name: NSNotification.Name("RefreshDashboard"), object: nil)
            }
        } catch {
            print("🔴 Failed to accept job: \(error)")
            throw error
        }
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

    func deleteJob(_ jobId: Int) async throws {
        print("🔵 JobService.deleteJob - Deleting job ID: \(jobId)")

        do {
            let response = try await apiService.deleteJob(jobID: jobId)
            print("🟢 Job deletion successful: \(response)")

            // Remove job from all local arrays
            self.jobs.removeAll { $0.id == jobId }
            self.myJobs.removeAll { $0.id == jobId }
            self.availableJobs.removeAll { $0.id == jobId }

            print("🟢 Job removed from local lists")

            // Notify dashboard to refresh
            NotificationCenter.default.post(name: NSNotification.Name("RefreshDashboard"), object: nil)
        } catch {
            print("🔴 Failed to delete job: \(error)")
            throw error
        }
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
        case customerId = "consumer_id"
        case gigworkerId = "gig_worker_id"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
        case scheduledFor = "scheduled_start"
    }

    // Memberwise initializer for programmatic creation
    init(id: Int?, uuid: String?, title: String, description: String, category: String?, location: String?, price: Double?, status: String?, customerId: Int?, gigworkerId: Int?, createdAt: String?, updatedAt: String?, scheduledFor: String?) {
        self.id = id
        self.uuid = uuid
        self.title = title
        self.description = description
        self.category = category
        self.location = location
        self.price = price
        self.status = status
        self.customerId = customerId
        self.gigworkerId = gigworkerId
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.scheduledFor = scheduledFor
    }

    // Decoder initializer for JSON decoding
    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = try container.decodeIfPresent(Int.self, forKey: .id)
        uuid = try container.decodeIfPresent(String.self, forKey: .uuid)
        title = try container.decode(String.self, forKey: .title)
        description = try container.decode(String.self, forKey: .description)
        category = try container.decodeIfPresent(String.self, forKey: .category)
        location = try container.decodeIfPresent(String.self, forKey: .location)
        price = try container.decodeIfPresent(Double.self, forKey: .price)
        status = try container.decodeIfPresent(String.self, forKey: .status)
        customerId = try container.decodeIfPresent(Int.self, forKey: .customerId)
        gigworkerId = try container.decodeIfPresent(Int.self, forKey: .gigworkerId)
        createdAt = try container.decodeIfPresent(String.self, forKey: .createdAt)
        updatedAt = try container.decodeIfPresent(String.self, forKey: .updatedAt)
        scheduledFor = try container.decodeIfPresent(String.self, forKey: .scheduledFor)

        print("🔵 Job decoded - title: \(title), status: \(status ?? "nil"), gigworkerId: \(gigworkerId?.description ?? "nil")")
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

