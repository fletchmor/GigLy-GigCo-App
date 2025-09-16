//
//  CreateJobView.swift
//  GigCo-Mobile
//
//  Created by Claude on 9/9/25.
//

import SwiftUI

struct CreateJobView: View {
    @Environment(\.dismiss) private var dismiss
    @EnvironmentObject var authService: AuthService
    @StateObject private var jobService = JobService()
    
    @State private var title = ""
    @State private var description = ""
    @State private var category = "General"
    @State private var location = ""
    @State private var price = ""
    @State private var scheduledDate = Date()
    @State private var hasScheduledDate = false
    
    @State private var isLoading = false
    @State private var showError = false
    @State private var errorMessage = ""
    
    private let categories = [
        "General", "Home Repair", "Cleaning", "Moving", "Delivery",
        "Tutoring", "Pet Care", "Yard Work", "Event Help", "Tech Support"
    ]
    
    var body: some View {
        NavigationView {
            Form {
                Section(header: Text("Job Details")) {
                    TextField("Job Title", text: $title)
                    
                    Picker("Category", selection: $category) {
                        ForEach(categories, id: \.self) { category in
                            Text(category).tag(category)
                        }
                    }
                    
                    TextField("Location", text: $location)
                    
                    TextField("Price ($)", text: $price)
                        .keyboardType(.decimalPad)
                }
                
                Section(header: Text("Description")) {
                    TextEditor(text: $description)
                        .frame(minHeight: 100)
                }
                
                Section(header: Text("Scheduling")) {
                    Toggle("Schedule for specific time", isOn: $hasScheduledDate)
                    
                    if hasScheduledDate {
                        DatePicker("Scheduled Date", selection: $scheduledDate, displayedComponents: [.date, .hourAndMinute])
                    }
                }
                
                Section {
                    Button(action: createJob) {
                        HStack {
                            if isLoading {
                                ProgressView()
                                    .progressViewStyle(CircularProgressViewStyle())
                                    .scaleEffect(0.8)
                            }
                            Text("Create Job")
                                .fontWeight(.semibold)
                        }
                    }
                    .frame(maxWidth: .infinity)
                    .disabled(!isFormValid || isLoading)
                    .buttonStyle(BorderlessButtonStyle())
                    .foregroundColor(isFormValid ? .blue : .gray)
                }
            }
            .navigationTitle("Create Job")
            .navigationBarTitleDisplayMode(.inline)
            .navigationBarBackButtonHidden(true)
            .toolbar {
                ToolbarItem(placement: .navigationBarLeading) {
                    Button("Cancel") {
                        dismiss()
                    }
                }
            }
            .alert("Error", isPresented: $showError) {
                Button("OK") { }
            } message: {
                Text(errorMessage)
            }
        }
    }
    
    private var isFormValid: Bool {
        !title.isEmpty &&
        !description.isEmpty &&
        !location.isEmpty &&
        !price.isEmpty &&
        Double(price) != nil
    }
    
    private func createJob() {
        guard isFormValid else {
            print("🔴 CreateJobView - Form validation failed")
            return
        }
        guard let priceValue = Double(price) else {
            print("🔴 CreateJobView - Invalid price value: \(price)")
            return
        }
        guard let currentUser = authService.currentUser,
              let consumerID = currentUser.id else {
            print("🔴 CreateJobView - User not authenticated or missing user ID")
            errorMessage = "User not authenticated or missing user ID"
            showError = true
            return
        }

        print("🔵 CreateJobView - Starting job creation")
        print("🔵 CreateJobView - User: \(currentUser.name), ID: \(consumerID)")
        print("🔵 CreateJobView - Job: \(title), Price: \(priceValue)")

        isLoading = true

        let scheduledFor = hasScheduledDate ? ISO8601DateFormatter().string(from: scheduledDate) : nil

        let jobRequest = CreateJobRequest(
            title: title,
            description: description,
            category: category,
            location: location,
            price: priceValue,
            scheduledFor: scheduledFor
        )

        print("🔵 CreateJobView - Job request created: \(jobRequest)")

        Task {
            do {
                let response = try await jobService.createJob(jobRequest, consumerID: consumerID)
                print("🟢 CreateJobView - Job created successfully: \(response)")
                await MainActor.run {
                    isLoading = false
                    dismiss()
                }
            } catch {
                print("🔴 CreateJobView - Job creation failed: \(error)")
                await MainActor.run {
                    errorMessage = error.localizedDescription
                    showError = true
                    isLoading = false
                }
            }
        }
    }
}