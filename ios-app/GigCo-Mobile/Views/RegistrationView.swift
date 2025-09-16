//
//  RegistrationView.swift
//  GigCo-Mobile
//
//  Created by Fletcher Morris on 9/8/25.
//

import SwiftUI

struct RegistrationView: View {
    @Environment(\.dismiss) private var dismiss
    @EnvironmentObject var authService: AuthService
    
    @State private var email = ""
    @State private var password = ""
    @State private var confirmPassword = ""
    @State private var firstName = ""
    @State private var lastName = ""
    @State private var address = ""
    @State private var selectedRole: PersonRole = .consumer
    @State private var isLoading = false
    @State private var showError = false
    @State private var errorMessage = ""
    
    var body: some View {
        NavigationView {
            Form {
                Section(header: Text("Personal Information")) {
                    TextField("First Name", text: $firstName)
                    TextField("Last Name", text: $lastName)
                    TextField("Email", text: $email)
                        .autocapitalization(.none)
                        .keyboardType(.emailAddress)
                    TextField("Address (Optional)", text: $address)
                }
                
                Section(header: Text("Security")) {
                    SecureField("Password", text: $password)
                    SecureField("Confirm Password", text: $confirmPassword)
                }
                
                Section(header: Text("Account Type")) {
                    Picker("I want to:", selection: $selectedRole) {
                        Text("Find services (Consumer)").tag(PersonRole.consumer)
                        Text("Offer services (Gig Worker)").tag(PersonRole.gigWorker)
                    }
                    .pickerStyle(SegmentedPickerStyle())
                }
                
                Section {
                    Button(action: {
                        print("🔵 Create Account button tapped")
                        register()
                    }) {
                        HStack {
                            if isLoading {
                                ProgressView()
                                    .progressViewStyle(CircularProgressViewStyle())
                                    .scaleEffect(0.8)
                            }
                            Text("Create Account")
                                .fontWeight(.semibold)
                        }
                    }
                    .frame(maxWidth: .infinity)
                    .disabled(!isFormValid || isLoading)
                    .buttonStyle(BorderlessButtonStyle())
                    .foregroundColor(isFormValid ? .blue : .gray)
                }
            }
            .navigationTitle("Create Account")
            .navigationBarTitleDisplayMode(.large)
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
        let valid = !firstName.isEmpty &&
        !lastName.isEmpty &&
        !email.isEmpty &&
        !password.isEmpty &&
        password == confirmPassword &&
        password.count >= 6
        
        print("🔍 Form validation - firstName: '\(firstName)', lastName: '\(lastName)', email: '\(email)', password: '\(password)', confirmPassword: '\(confirmPassword)', valid: \(valid)")
        return valid
    }
    
    private func register() {
        guard isFormValid else { 
            print("🔴 Form validation failed")
            return 
        }
        
        print("🔵 Starting registration process...")
        isLoading = true
        
        Task {
            do {
                print("🔵 Calling authService.register...")
                try await authService.register(
                    email: email,
                    password: password,
                    firstName: firstName,
                    lastName: lastName,
                    role: selectedRole,
                    address: address
                )
                print("🟢 Registration successful, dismissing view")
                await MainActor.run {
                    isLoading = false
                    dismiss()
                }
            } catch {
                print("🔴 Registration failed: \(error)")
                await MainActor.run {
                    errorMessage = error.localizedDescription
                    showError = true
                    isLoading = false
                }
            }
        }
    }
}