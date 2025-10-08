//
//  JobRowView.swift
//  GigCo-Mobile
//
//  Created by Claude on 9/11/25.
//

import SwiftUI

struct JobRowView: View {
    let job: Job
    let jobService: JobService
    @EnvironmentObject var authService: AuthService
    @State private var isAccepting = false
    @State private var showError = false
    @State private var errorMessage = ""
    @State private var showDeleteConfirmation = false
    @State private var isDeleting = false

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    Text(job.title)
                        .font(.headline)
                        .fontWeight(.semibold)
                    
                    Text(job.description)
                        .font(.subheadline)
                        .foregroundColor(.secondary)
                        .lineLimit(2)
                }
                
                Spacer()
                
                VStack(alignment: .trailing, spacing: 4) {
                    if let price = job.price {
                        Text("$\(String(format: "%.2f", price))")
                            .font(.headline)
                            .fontWeight(.bold)
                            .foregroundColor(.green)
                    }
                    
                    if let status = job.status {
                        Text(status.capitalized)
                            .font(.caption)
                            .padding(.horizontal, 8)
                            .padding(.vertical, 4)
                            .background(statusColor(for: status))
                            .foregroundColor(.white)
                            .cornerRadius(12)
                    }
                }
            }
            
            HStack {
                if let location = job.location {
                    HStack(spacing: 4) {
                        Image(systemName: "location")
                            .foregroundColor(.secondary)
                            .font(.caption)
                        Text(location)
                            .font(.caption)
                            .foregroundColor(.secondary)
                    }
                }
                
                Spacer()
                
                if let category = job.category {
                    Text(category)
                        .font(.caption)
                        .padding(.horizontal, 6)
                        .padding(.vertical, 2)
                        .background(Color.blue.opacity(0.1))
                        .foregroundColor(.blue)
                        .cornerRadius(6)
                }
            }

            // Action buttons
            if shouldShowAcceptButton || shouldShowDeleteButton {
                HStack {
                    Spacer()

                    // Add Accept Job button for workers viewing available jobs
                    if shouldShowAcceptButton {
                        Button(action: acceptJob) {
                            HStack {
                                if isAccepting {
                                    ProgressView()
                                        .progressViewStyle(CircularProgressViewStyle(tint: .white))
                                        .scaleEffect(0.8)
                                    Text("Accepting...")
                                } else {
                                    Text("Accept Job")
                                }
                            }
                            .font(.subheadline)
                            .fontWeight(.semibold)
                            .foregroundColor(.white)
                            .padding(.horizontal, 16)
                            .padding(.vertical, 8)
                            .background(Color.blue)
                            .cornerRadius(8)
                        }
                        .disabled(isAccepting)
                    }

                    // Add spacing between buttons if both are showing
                    if shouldShowAcceptButton && shouldShowDeleteButton {
                        Spacer().frame(width: 8)
                    }

                    // Add Delete Job button for job creators
                    if shouldShowDeleteButton {
                        Button(action: {
                            print("🔵 JobRowView - Delete button tapped for job: \(job.title)")
                            showDeleteConfirmation = true
                        }) {
                            HStack(spacing: 4) {
                                if isDeleting {
                                    ProgressView()
                                        .progressViewStyle(CircularProgressViewStyle(tint: .white))
                                        .scaleEffect(0.8)
                                    Text("Deleting...")
                                } else {
                                    Image(systemName: "trash")
                                    Text("Delete")
                                }
                            }
                            .font(.subheadline)
                            .fontWeight(.semibold)
                            .foregroundColor(.white)
                            .padding(.horizontal, 20)
                            .padding(.vertical, 12)
                            .background(Color.red)
                            .cornerRadius(8)
                        }
                        .disabled(isDeleting)
                        .buttonStyle(.plain)
                        .contentShape(Rectangle()) // Make entire button area tappable
                        .onTapGesture {
                            print("🟡 JobRowView - Alternative tap gesture triggered")
                            showDeleteConfirmation = true
                        }
                    }
                }
                .padding(.top, 8)
            }
        }
        .padding(.vertical, 4)
        .alert("Error", isPresented: $showError) {
            Button("OK") { }
        } message: {
            Text(errorMessage)
        }
        .confirmationDialog("Delete Job", isPresented: $showDeleteConfirmation) {
            Button("Delete", role: .destructive) {
                print("🔵 JobRowView - Delete confirmed, calling deleteJob()")
                deleteJob()
            }
            Button("Cancel", role: .cancel) {
                print("🔵 JobRowView - Delete cancelled")
            }
        } message: {
            Text("Are you sure you want to permanently delete this job? This action cannot be undone.")
        }
    }

    private var shouldShowAcceptButton: Bool {
        // Show accept button if:
        // 1. Current user is a gig worker
        // 2. Job status is "posted" or "available"
        // 3. Job is not already assigned to someone

        print("🔵 JobRowView.shouldShowAcceptButton - Checking accept button visibility")
        print("🔵 JobRowView - Current user: \(authService.currentUser?.name ?? "nil"), Role: \(authService.currentUser?.role ?? "nil")")
        print("🔵 JobRowView - Job: \(job.title), Status: \(job.status ?? "nil"), gigworkerId: \(job.gigworkerId?.description ?? "nil")")

        guard let currentUser = authService.currentUser,
              currentUser.role == "gig_worker",
              let status = job.status,
              (status == "posted" || status == "available"),
              job.gigworkerId == nil else {
            print("🔴 JobRowView - Accept button NOT shown")
            return false
        }

        print("🟢 JobRowView - Accept button SHOULD be shown")
        return true
    }

    private var shouldShowDeleteButton: Bool {
        // Show delete button if:
        // 1. Current user is the job creator (consumer)
        // 2. Job status is "posted" or "cancelled" (not in progress or completed)

        print("🔵 JobRowView.shouldShowDeleteButton - Checking delete button visibility")
        print("🔵 JobRowView - Current user: \(authService.currentUser?.name ?? "nil"), ID: \(authService.currentUser?.id ?? -1)")
        print("🔵 JobRowView - Job: \(job.title), Creator ID: \(job.customerId ?? -1), Status: \(job.status ?? "nil")")

        guard let currentUser = authService.currentUser,
              let currentUserId = currentUser.id,
              let jobCreatorId = job.customerId,
              currentUserId == jobCreatorId,
              let status = job.status,
              (status == "posted" || status == "cancelled") else {
            print("🔴 JobRowView - Delete button NOT shown")
            return false
        }

        print("🟢 JobRowView - Delete button SHOULD be shown")
        return true
    }

    private func acceptJob() {
        print("🔵 JobRowView.acceptJob - Button tapped!")
        print("🔵 JobRowView.acceptJob - Current user: \(authService.currentUser?.name ?? "nil")")
        print("🔵 JobRowView.acceptJob - Job ID: \(job.id?.description ?? "nil")")
        print("🔵 JobRowView.acceptJob - User ID: \(authService.currentUser?.id?.description ?? "nil")")

        guard let currentUser = authService.currentUser,
              let jobId = job.id,
              let gigWorkerId = currentUser.id else {
            print("🔴 JobRowView.acceptJob - Missing required data")
            print("🔴 JobRowView.acceptJob - currentUser: \(authService.currentUser != nil)")
            print("🔴 JobRowView.acceptJob - jobId: \(job.id != nil)")
            print("🔴 JobRowView.acceptJob - gigWorkerId: \(authService.currentUser?.id != nil)")
            errorMessage = "Unable to accept job. Please try again."
            showError = true
            return
        }

        print("🔵 JobRowView.acceptJob - Starting acceptance for job \(jobId) by worker \(gigWorkerId)")
        isAccepting = true

        Task {
            do {
                print("🔵 JobRowView.acceptJob - Calling jobService.acceptJob")
                try await jobService.acceptJob(jobId, gigWorkerID: gigWorkerId)
                await MainActor.run {
                    print("🟢 JobRowView.acceptJob - Success!")
                    isAccepting = false
                }
                // Job will be updated in the service automatically
            } catch {
                await MainActor.run {
                    print("🔴 JobRowView.acceptJob - Error: \(error)")
                    isAccepting = false
                    errorMessage = error.localizedDescription
                    showError = true
                }
            }
        }
    }

    private func deleteJob() {
        print("🔵 JobRowView.deleteJob - Starting delete process")

        guard let jobId = job.id else {
            print("🔴 JobRowView.deleteJob - No job ID found")
            errorMessage = "Unable to delete job. Please try again."
            showError = true
            return
        }

        print("🔵 JobRowView.deleteJob - Deleting job with ID: \(jobId)")
        isDeleting = true

        Task {
            do {
                print("🔵 JobRowView.deleteJob - Calling jobService.deleteJob")
                try await jobService.deleteJob(jobId)
                await MainActor.run {
                    print("🟢 JobRowView.deleteJob - Delete successful")
                    isDeleting = false
                }
                // Job will be removed from the lists automatically in JobService
            } catch {
                await MainActor.run {
                    print("🔴 JobRowView.deleteJob - Delete failed: \(error)")
                    isDeleting = false
                    errorMessage = error.localizedDescription
                    showError = true
                }
            }
        }
    }

    private func statusColor(for status: String) -> Color {
        switch status.lowercased() {
        case "posted":
            return .blue
        case "accepted":
            return .cyan
        case "in_progress":
            return .orange
        case "completed":
            return .green
        case "cancelled":
            return .red
        default:
            return .gray
        }
    }
}

#Preview {
    JobRowView(
        job: Job(
            id: 1,
            uuid: "123",
            title: "House Cleaning",
            description: "Need someone to clean my house thoroughly",
            category: "Cleaning",
            location: "San Francisco, CA",
            price: 75.0,
            status: "open",
            customerId: 1,
            gigworkerId: nil,
            createdAt: "2025-01-01",
            updatedAt: "2025-01-01",
            scheduledFor: nil
        ),
        jobService: JobService()
    )
    .padding()
}