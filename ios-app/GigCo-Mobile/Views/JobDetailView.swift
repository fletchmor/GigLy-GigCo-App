//
//  JobDetailView.swift
//  GigCo-Mobile
//
//  Created by Claude on 9/9/25.
//

import SwiftUI

struct JobDetailView: View {
    let job: Job
    @EnvironmentObject var authService: AuthService
    @StateObject private var jobService = JobService()
    @Environment(\.dismiss) private var dismiss
    
    @State private var showingAcceptAlert = false
    @State private var showError = false
    @State private var errorMessage = ""
    @State private var showSuccess = false
    @State private var successMessage = ""
    
    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 20) {
                // Header Section
                VStack(alignment: .leading, spacing: 12) {
                    HStack {
                        VStack(alignment: .leading, spacing: 4) {
                            Text(job.title)
                                .font(.title2)
                                .fontWeight(.bold)
                            
                            Text(job.category ?? "General")
                                .font(.caption)
                                .padding(.horizontal, 8)
                                .padding(.vertical, 2)
                                .background(Color.blue.opacity(0.1))
                                .foregroundColor(.blue)
                                .cornerRadius(4)
                        }
                        
                        Spacer()
                        
                        VStack(alignment: .trailing) {
                            if let price = job.price {
                                Text("$\(String(format: "%.2f", price))")
                                    .font(.title2)
                                    .fontWeight(.bold)
                                    .foregroundColor(.green)
                            }
                            
                            Text(job.status?.capitalized ?? "Unknown")
                                .font(.caption)
                                .foregroundColor(statusColor)
                        }
                    }
                }
                .padding()
                .background(Color(.systemGray6))
                .cornerRadius(12)
                
                // Job Information
                VStack(alignment: .leading, spacing: 16) {
                    // Posted by section
                    if let consumerName = job.consumerName {
                        VStack(alignment: .leading, spacing: 8) {
                            Text("Posted By")
                                .font(.headline)

                            HStack {
                                Image(systemName: "person.circle.fill")
                                    .foregroundColor(.blue)
                                    .font(.title3)
                                Text(consumerName)
                                    .font(.body)
                            }
                        }
                    }

                    VStack(alignment: .leading, spacing: 8) {
                        Text("Description")
                            .font(.headline)

                        Text(job.description)
                            .font(.body)
                    }
                    
                    if let location = job.location {
                        VStack(alignment: .leading, spacing: 8) {
                            Text("Location")
                                .font(.headline)
                            
                            HStack {
                                Image(systemName: "location")
                                    .foregroundColor(.blue)
                                Text(location)
                            }
                        }
                    }
                    
                    if let scheduledFor = job.scheduledFor {
                        VStack(alignment: .leading, spacing: 8) {
                            Text("Scheduled For")
                                .font(.headline)
                            
                            HStack {
                                Image(systemName: "calendar")
                                    .foregroundColor(.blue)
                                Text(formatDate(scheduledFor))
                            }
                        }
                    }
                }
                .padding()
                .background(Color.white)
                .cornerRadius(12)
                .shadow(radius: 1)
                
                // Reviews Section - TODO: Implement when review system is available
                // Placeholder for future reviews functionality
                
                // Debug info
                if let status = job.status, let role = authService.currentUser?.role {
                    Text("Status: \(status) | Role: \(role)")
                        .font(.caption)
                        .foregroundColor(.secondary)
                        .padding(.horizontal)
                }

                // Action Buttons
                VStack(spacing: 12) {
                    if canAcceptJob {
                        Button("Accept Job") {
                            showingAcceptAlert = true
                        }
                        .buttonStyle(.borderedProminent)
                        .controlSize(.large)
                        .frame(maxWidth: .infinity)
                    }

                    if canStartJob {
                        Button("Start Job") {
                            print("🔵 JobDetailView - Start Job button tapped")
                            startJob()
                        }
                        .buttonStyle(.borderedProminent)
                        .controlSize(.large)
                        .frame(maxWidth: .infinity)
                    }

                    if canCompleteJob {
                        Button("Complete Job") {
                            print("🔵 JobDetailView - Complete Job button tapped")
                            completeJob()
                        }
                        .buttonStyle(.borderedProminent)
                        .controlSize(.large)
                        .frame(maxWidth: .infinity)
                    }

                    // Show helpful message if no action buttons available
                    if !canAcceptJob && !canStartJob && !canCompleteJob && !canCancelJob {
                        Text("No actions available for this job")
                            .font(.subheadline)
                            .foregroundColor(.secondary)
                            .padding()
                    }
                    
                    // Review functionality temporarily removed - TODO: Implement when review system is available
                    
                    if canCancelJob {
                        Button("Cancel Job") {
                            cancelJob()
                        }
                        .buttonStyle(.bordered)
                        .foregroundColor(.red)
                        .controlSize(.large)
                        .frame(maxWidth: .infinity)
                    }
                }
                .padding()
                
                Spacer()
            }
            .padding()
        }
        .navigationTitle("Job Details")
        .navigationBarTitleDisplayMode(.inline)
        .alert("Accept Job", isPresented: $showingAcceptAlert) {
            Button("Cancel", role: .cancel) { }
            Button("Accept") {
                acceptJob()
            }
        } message: {
            Text("Are you sure you want to accept this job?")
        }
        .alert("Error", isPresented: $showError) {
            Button("OK") { }
        } message: {
            Text(errorMessage)
        }
        .alert("Success", isPresented: $showSuccess) {
            Button("OK") {
                dismiss()
            }
        } message: {
            Text(successMessage)
        }
    }
    
    // MARK: - Computed Properties
    
    private var statusColor: Color {
        switch job.status?.lowercased() {
        case "posted":
            return .green
        case "accepted":
            return .blue
        case "in_progress":
            return .orange
        case "completed":
            return .purple
        case "cancelled":
            return .red
        default:
            return .secondary
        }
    }
    
    private var canAcceptJob: Bool {
        let result = job.status == "posted" && authService.currentUser?.role == "gig_worker"
        print("🔵 JobDetailView - canAcceptJob: \(result) (status: \(job.status ?? "nil"), role: \(authService.currentUser?.role ?? "nil"))")
        return result
    }

    private var canStartJob: Bool {
        let result = job.status == "accepted" && authService.currentUser?.role == "gig_worker"
        print("🔵 JobDetailView - canStartJob: \(result) (status: \(job.status ?? "nil"), role: \(authService.currentUser?.role ?? "nil"))")
        return result
    }

    private var canCompleteJob: Bool {
        // Both worker and consumer can mark job complete when accepted, in_progress, or completed (for dual confirmation)
        // Workers can complete from accepted status (skipping start if they want)
        // Consumers can confirm completion when in_progress or completed
        let userRole = authService.currentUser?.role

        if userRole == "gig_worker" {
            // Workers can complete from accepted, in_progress, or completed status
            let canComplete = job.status == "accepted" || job.status == "in_progress" || job.status == "completed"
            print("🔵 JobDetailView - canCompleteJob (worker): \(canComplete) (status: \(job.status ?? "nil"))")
            return canComplete
        } else if userRole == "consumer" {
            // Consumers can only confirm when job is in_progress or completed
            let canComplete = job.status == "in_progress" || job.status == "completed"
            print("🔵 JobDetailView - canCompleteJob (consumer): \(canComplete) (status: \(job.status ?? "nil"))")
            return canComplete
        }

        print("🔵 JobDetailView - canCompleteJob: false (invalid role or status)")
        return false
    }
    
    
    private var canCancelJob: Bool {
        (job.status == "posted" || job.status == "accepted") &&
        authService.currentUser?.role == "consumer"
    }
    
    // MARK: - Helper Functions
    
    private func formatDate(_ dateString: String) -> String {
        let formatter = ISO8601DateFormatter()
        if let date = formatter.date(from: dateString) {
            let displayFormatter = DateFormatter()
            displayFormatter.dateStyle = .medium
            displayFormatter.timeStyle = .short
            return displayFormatter.string(from: date)
        }
        return dateString
    }
    
    
    // MARK: - Job Actions
    
    private func acceptJob() {
        guard let jobId = job.id,
              let currentUser = authService.currentUser,
              let gigWorkerId = currentUser.id else {
            errorMessage = "Unable to accept job. Please try again."
            showError = true
            return
        }

        Task {
            do {
                try await jobService.acceptJob(jobId, gigWorkerID: gigWorkerId)
                await MainActor.run {
                    dismiss()
                }
            } catch {
                await MainActor.run {
                    errorMessage = error.localizedDescription
                    showError = true
                }
            }
        }
    }
    
    private func startJob() {
        guard let jobId = job.id else { return }

        Task {
            do {
                try await jobService.startJob(jobId)
                await MainActor.run {
                    successMessage = "Job started successfully! You can now complete it when finished."
                    showSuccess = true
                }
            } catch {
                await MainActor.run {
                    errorMessage = error.localizedDescription
                    showError = true
                }
            }
        }
    }

    private func completeJob() {
        guard let jobId = job.id else { return }

        Task {
            do {
                try await jobService.completeJob(jobId)
                await MainActor.run {
                    successMessage = "Job completion confirmed! The other party will also need to confirm."
                    showSuccess = true
                }
            } catch {
                await MainActor.run {
                    errorMessage = error.localizedDescription
                    showError = true
                }
            }
        }
    }
    
    private func cancelJob() {
        guard let jobId = job.id else { return }
        
        Task {
            do {
                try await jobService.cancelJob(jobId)
                await MainActor.run {
                    dismiss()
                }
            } catch {
                await MainActor.run {
                    errorMessage = error.localizedDescription
                    showError = true
                }
            }
        }
    }
}