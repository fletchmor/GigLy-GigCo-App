//
//  LoginView.swift
//  GigCo-Mobile
//
//  Created by Fletcher Morris on 9/8/25.
//

import SwiftUI

struct LoginView: View {
    @EnvironmentObject var authService: AuthService
    @State private var email = ""
    @State private var password = ""
    @State private var isLoading = false
    @State private var showError = false
    @State private var errorMessage = ""
    @State private var showRegistration = false
    @State private var showPassword = false
    
    var body: some View {
        NavigationView {
            VStack(spacing: 30) {
                // Logo/Header
                VStack {
                    Text("GigCo")
                        .font(.largeTitle)
                        .fontWeight(.bold)
                        .foregroundColor(.blue)
                    Text("Connect. Work. Earn.")
                        .font(.subheadline)
                        .foregroundColor(.gray)
                }
                .padding(.top, 50)
                
                // Login Form
                VStack(spacing: 20) {
                    TextField("Email", text: $email)
                        .textFieldStyle(RoundedBorderTextFieldStyle())
                        .autocapitalization(.none)
                        .keyboardType(.emailAddress)
                        .frame(height: 44)
                    
                    HStack {
                        if showPassword {
                            TextField("Password", text: $password)
                                .textInputAutocapitalization(.never)
                                .autocorrectionDisabled(true)
                        } else {
                            SecureField("Password", text: $password)
                        }

                        Button(action: {
                            showPassword.toggle()
                        }) {
                            Image(systemName: showPassword ? "eye.slash" : "eye")
                                .foregroundColor(.gray)
                        }
                    }
                    .textFieldStyle(RoundedBorderTextFieldStyle())
                    .frame(height: 44)
                    
                    Button(action: login) {
                        HStack {
                            if isLoading {
                                ProgressView()
                                    .progressViewStyle(CircularProgressViewStyle(tint: .white))
                                    .scaleEffect(0.8)
                            }
                            Text("Login")
                        }
                        .frame(maxWidth: .infinity)
                        .frame(height: 50)
                        .padding()
                        .background(Color.blue)
                        .foregroundColor(.white)
                        .cornerRadius(8)
                    }
                    .disabled(isLoading || email.isEmpty || password.isEmpty)
                }
                .padding(.horizontal)
                
                // Registration Link
                Button(action: { showRegistration = true }) {
                    Text("Don't have an account? Sign Up")
                        .foregroundColor(.blue)
                }
                
                Spacer()
            }
            .navigationBarHidden(true)
            .onTapGesture {
                UIApplication.shared.sendAction(#selector(UIResponder.resignFirstResponder), to: nil, from: nil, for: nil)
            }
            .alert("Error", isPresented: $showError) {
                Button("OK") { }
            } message: {
                Text(errorMessage)
            }
            .sheet(isPresented: $showRegistration) {
                RegistrationView()
                    .environmentObject(authService)
            }
        }
    }
    
    private func login() {
        guard !email.isEmpty, !password.isEmpty else { return }
        
        isLoading = true
        
        Task {
            do {
                try await authService.login(email: email, password: password)
                await MainActor.run {
                    isLoading = false
                }
                // Navigation will be handled by the parent view observing authService.isAuthenticated
            } catch {
                await MainActor.run {
                    errorMessage = error.localizedDescription
                    showError = true
                    isLoading = false
                }
            }
        }
    }
}