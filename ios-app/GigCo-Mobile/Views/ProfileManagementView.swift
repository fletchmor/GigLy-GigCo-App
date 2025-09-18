//
//  ProfileManagementView.swift
//  GigCo-Mobile
//
//  Created by Claude on 9/9/25.
//

import SwiftUI

struct ProfileManagementView: View {
    @EnvironmentObject var authService: AuthService
    @EnvironmentObject var userService: UserService
    
    @State private var isEditing = false
    @State private var isLoading = false
    @State private var showError = false
    @State private var errorMessage = ""
    @State private var showGigWorkerSetup = false
    
    // Editable fields
    @State private var name = ""
    @State private var email = ""
    @State private var phoneNumber = ""
    @State private var address = ""
    
    var body: some View {
        NavigationView {
            Form {
                Section(header: Text("Profile Information")) {
                    HStack {
                        Circle()
                            .fill(Color.blue.opacity(0.2))
                            .frame(width: 60, height: 60)
                            .overlay(
                                Text(userInitials)
                                    .font(.title2)
                                    .fontWeight(.bold)
                                    .foregroundColor(.blue)
                            )
                        
                        VStack(alignment: .leading, spacing: 4) {
                            if isEditing {
                                TextField("Name", text: $name)
                                    .textFieldStyle(RoundedBorderTextFieldStyle())
                            } else {
                                Text(authService.currentUser?.name ?? "Unknown User")
                                    .font(.title2)
                                    .fontWeight(.semibold)
                            }
                            
                            Text(authService.currentUser?.role.capitalized ?? "Unknown")
                                .font(.caption)
                                .padding(.horizontal, 8)
                                .padding(.vertical, 2)
                                .background(Color.blue.opacity(0.1))
                                .foregroundColor(.blue)
                                .cornerRadius(8)
                        }
                        
                        Spacer()
                    }
                    .padding(.vertical, 8)
                }
                
                Section(header: Text("Contact Information")) {
                    HStack {
                        Image(systemName: "envelope")
                            .foregroundColor(.blue)
                            .frame(width: 24)
                        
                        if isEditing {
                            TextField("Email", text: $email)
                                .autocapitalization(.none)
                                .keyboardType(.emailAddress)
                        } else {
                            Text(authService.currentUser?.email ?? "No email")
                        }
                    }
                    
                    HStack {
                        Image(systemName: "phone")
                            .foregroundColor(.blue)
                            .frame(width: 24)
                        
                        if isEditing {
                            TextField("Phone Number", text: $phoneNumber)
                                .keyboardType(.phonePad)
                        } else {
                            Text(phoneNumber.isEmpty ? "No phone number" : phoneNumber)
                        }
                    }
                    
                    HStack {
                        Image(systemName: "location")
                            .foregroundColor(.blue)
                            .frame(width: 24)
                        
                        if isEditing {
                            TextField("Address", text: $address)
                        } else {
                            Text(address.isEmpty ? "No address" : address)
                        }
                    }
                }
                
                if authService.currentUser?.role == "consumer" {
                    Section(header: Text("Gig Worker Options")) {
                        Button("Become a Gig Worker") {
                            showGigWorkerSetup = true
                        }
                        .foregroundColor(.blue)
                    }
                }
                
                Section(header: Text("Account Management")) {
                    if isEditing {
                        Button("Save Changes") {
                            saveProfile()
                        }
                        .foregroundColor(.blue)
                        .disabled(isLoading)
                        
                        Button("Cancel") {
                            cancelEditing()
                        }
                        .foregroundColor(.red)
                    } else {
                        Button("Edit Profile") {
                            startEditing()
                        }
                        .foregroundColor(.blue)
                    }
                    
                    Button("Settings") {
                        // Navigate to settings
                    }
                    .foregroundColor(.blue)
                    
                    Button("Help & Support") {
                        // Navigate to help
                    }
                    .foregroundColor(.blue)
                }
                
                Section {
                    Button("Logout") {
                        authService.logout()
                    }
                    .foregroundColor(.red)
                    .frame(maxWidth: .infinity)
                }
            }
            .navigationTitle("Profile")
            .alert("Error", isPresented: $showError) {
                Button("OK") { }
            } message: {
                Text(errorMessage)
            }
            .sheet(isPresented: $showGigWorkerSetup) {
                GigWorkerSetupView()
                    .environmentObject(authService)
            }
            .task {
                loadProfile()
            }
        }
    }
    
    private var userInitials: String {
        guard let user = authService.currentUser else { return "U" }
        let nameComponents = user.name.components(separatedBy: " ")
        let firstInitial = nameComponents.first?.prefix(1).uppercased() ?? "U"
        let lastInitial = nameComponents.count > 1 ? nameComponents.last?.prefix(1).uppercased() ?? "" : ""
        return "\(firstInitial)\(lastInitial)"
    }
    
    private func loadProfile() {
        if let user = authService.currentUser {
            name = user.name
            email = user.email
        }
    }
    
    private func startEditing() {
        loadProfile()
        isEditing = true
    }
    
    private func cancelEditing() {
        loadProfile()
        isEditing = false
    }
    
    private func saveProfile() {
        isLoading = true
        
        Task {
            do {
                // TODO: Implement profile update when API endpoint is available
                print("🔵 Would update profile - name: \(name), email: \(email)")
                // Simulate API call delay
                try await Task.sleep(nanoseconds: 1_000_000_000) // 1 second
                print("🟢 Profile update simulated successfully")
                
                await MainActor.run {
                    isLoading = false
                    isEditing = false
                    // Optionally refresh the current user data
                }
            } catch {
                await MainActor.run {
                    isLoading = false
                    errorMessage = error.localizedDescription
                    showError = true
                }
            }
        }
    }
}

struct GigWorkerSetupView: View {
    @Environment(\.dismiss) private var dismiss
    @EnvironmentObject var authService: AuthService
    
    @State private var skills = ""
    @State private var experience = ""
    @State private var hourlyRate = ""
    @State private var availability = "flexible"
    @State private var bio = ""
    
    @State private var isLoading = false
    @State private var showError = false
    @State private var errorMessage = ""
    
    private let availabilityOptions = ["flexible", "weekdays", "weekends", "evenings"]
    
    var body: some View {
        NavigationView {
            Form {
                Section(header: Text("Skills & Experience")) {
                    TextField("Skills (comma separated)", text: $skills)
                    
                    Picker("Experience Level", selection: $experience) {
                        Text("Beginner").tag("beginner")
                        Text("Intermediate").tag("intermediate")
                        Text("Advanced").tag("advanced")
                        Text("Expert").tag("expert")
                    }
                    
                    TextField("Hourly Rate ($)", text: $hourlyRate)
                        .keyboardType(.decimalPad)
                }
                
                Section(header: Text("Availability")) {
                    Picker("Availability", selection: $availability) {
                        ForEach(availabilityOptions, id: \.self) { option in
                            Text(option.capitalized).tag(option)
                        }
                    }
                }
                
                Section(header: Text("About You")) {
                    TextEditor(text: $bio)
                        .frame(minHeight: 100)
                }
                
                Section {
                    Button(action: createGigWorkerProfile) {
                        HStack {
                            if isLoading {
                                ProgressView()
                                    .progressViewStyle(CircularProgressViewStyle())
                                    .scaleEffect(0.8)
                            }
                            Text("Become a Gig Worker")
                                .fontWeight(.semibold)
                        }
                    }
                    .frame(maxWidth: .infinity)
                    .disabled(!isFormValid || isLoading)
                    .buttonStyle(BorderlessButtonStyle())
                    .foregroundColor(isFormValid ? .blue : .gray)
                }
            }
            .navigationTitle("Become a Gig Worker")
            .navigationBarTitleDisplayMode(.inline)
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
        !skills.isEmpty &&
        !experience.isEmpty &&
        !bio.isEmpty
    }
    
    private func createGigWorkerProfile() {
        guard isFormValid else { return }
        guard let currentUser = authService.currentUser else {
            errorMessage = "User not authenticated"
            showError = true
            return
        }

        isLoading = true

        let _ = skills.split(separator: ",").map { $0.trimmingCharacters(in: .whitespaces) }
        let hourlyRateValue = Double(hourlyRate) ?? 20.0

        // Create bio with skills and experience
        let combinedBio = bio.isEmpty ? "Skills: \(skills)\nExperience: \(experience)" : "\(bio)\n\nSkills: \(skills)\nExperience: \(experience)"

        // Map experience level to years
        let experienceYears: Int
        switch experience {
        case "beginner":
            experienceYears = 1
        case "intermediate":
            experienceYears = 3
        case "advanced":
            experienceYears = 5
        case "expert":
            experienceYears = 10
        default:
            experienceYears = 1
        }

        Task {
            do {
                let userService = UserService.shared
                let response = try await userService.createGigWorkerProfile(
                    name: currentUser.name,
                    email: currentUser.email,
                    bio: combinedBio,
                    hourlyRate: hourlyRateValue,
                    experienceYears: experienceYears
                )

                print("🟢 GigWorkerSetupView - Gig worker profile created successfully: \(response)")

                await MainActor.run {
                    isLoading = false
                    dismiss()
                }
            } catch {
                print("🔴 GigWorkerSetupView - Gig worker profile creation failed: \(error)")
                await MainActor.run {
                    errorMessage = error.localizedDescription
                    showError = true
                    isLoading = false
                }
            }
        }
    }
}