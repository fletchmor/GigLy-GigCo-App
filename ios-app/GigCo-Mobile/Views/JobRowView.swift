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

            // Add Accept Job button for workers viewing available jobs
            if shouldShowAcceptButton {
                HStack {
                    Spacer()
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
                .padding(.top, 4)
            }
        }
        .padding(.vertical, 4)
        .alert("Error", isPresented: $showError) {
            Button("OK") { }
        } message: {
            Text(errorMessage)
        }
    }

    private var shouldShowAcceptButton: Bool {
        // Show accept button if:
        // 1. Current user is a gig worker
        // 2. Job status is "posted" or "available"
        // 3. Job is not already assigned to someone
        guard let currentUser = authService.currentUser,
              currentUser.role == "gig_worker",
              let status = job.status,
              (status == "posted" || status == "available"),
              job.gigworkerId == nil else {
            return false
        }
        return true
    }

    private func acceptJob() {
        guard let currentUser = authService.currentUser,
              let jobId = job.id,
              let gigWorkerId = currentUser.id else {
            errorMessage = "Unable to accept job. Please try again."
            showError = true
            return
        }

        isAccepting = true

        Task {
            do {
                try await jobService.acceptJob(jobId, gigWorkerID: gigWorkerId)
                await MainActor.run {
                    isAccepting = false
                }
                // Job will be updated in the service automatically
            } catch {
                await MainActor.run {
                    isAccepting = false
                    errorMessage = error.localizedDescription
                    showError = true
                }
            }
        }
    }
    
    private func statusColor(for status: String) -> Color {
        switch status.lowercased() {
        case "open", "available":
            return .blue
        case "in_progress", "accepted":
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