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
        }
        .padding(.vertical, 4)
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