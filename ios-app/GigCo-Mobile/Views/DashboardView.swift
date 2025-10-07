//
//  DashboardView.swift
//  GigCo-Mobile
//
//  Created by Fletcher Morris on 9/8/25.
//

import SwiftUI

struct DashboardView: View {
    @EnvironmentObject var authService: AuthService
    
    var body: some View {
        TabView {
            HomeTabView()
                .tabItem {
                    Image(systemName: "house.fill")
                    Text("Home")
                }
                .environmentObject(authService)
            
            JobListView()
                .tabItem {
                    Image(systemName: "briefcase.fill")
                    Text("Jobs")
                }
                .environmentObject(authService)
            
            ProfileManagementView()
                .tabItem {
                    Image(systemName: "person.fill")
                    Text("Profile")
                }
                .environmentObject(authService)
        }
        .accentColor(.blue)
    }
}

struct HomeTabView: View {
    @EnvironmentObject var authService: AuthService
    @EnvironmentObject var apiService: APIService
    @StateObject private var jobService = JobService()
    @State private var showingJobList = false
    @State private var showingCreateJob = false
    @State private var activeJobsCount = 0
    @State private var completedJobsCount = 0
    @State private var apiHealthy = false
    @State private var healthCheckMessage = "Checking API..."
    
    var body: some View {
        NavigationView {
            ScrollView {
                VStack(alignment: .leading, spacing: 20) {
                    // Welcome Section
                    VStack(alignment: .leading, spacing: 8) {
                        Text("Welcome to GigCo")
                            .font(.title2)
                            .fontWeight(.bold)
                        
                        HStack {
                            if let user = authService.currentUser {
                                Text("Hello, \(user.name.components(separatedBy: " ").first ?? "User")!")
                                    .font(.subheadline)
                                    .foregroundColor(.blue)
                            } else {
                                Text("Find services or offer your skills")
                                    .font(.subheadline)
                                    .foregroundColor(.gray)
                            }

                            Spacer()

                            Button("Logout") {
                                authService.logout()
                            }
                            .font(.caption)
                            .foregroundColor(.red)
                            .padding(.horizontal, 8)
                            .padding(.vertical, 4)
                            .background(Color.red.opacity(0.1))
                            .cornerRadius(8)
                        }
                    }
                    .padding(.horizontal)
                    
                    // API Status
                    HStack {
                        Circle()
                            .fill(apiHealthy ? .green : .red)
                            .frame(width: 8, height: 8)
                        Text(healthCheckMessage)
                            .font(.caption)
                            .foregroundColor(.secondary)
                        Spacer()
                    }
                    .padding(.horizontal)
                    
                    // Quick Actions
                    LazyVGrid(columns: [
                        GridItem(.flexible()),
                        GridItem(.flexible())
                    ], spacing: 16) {
                        if authService.currentUser?.role == "gig_worker" {
                            QuickActionCard(
                                title: "Find Work",
                                icon: "magnifyingglass",
                                color: .blue
                            ) {
                                showingJobList = true
                            }
                            
                            QuickActionCard(
                                title: "My Jobs",
                                icon: "briefcase",
                                color: .orange
                            ) {
                                // Will navigate to jobs tab
                            }
                        } else {
                            QuickActionCard(
                                title: "Browse Services",
                                icon: "magnifyingglass",
                                color: .blue
                            ) {
                                showingJobList = true
                            }
                            
                            QuickActionCard(
                                title: "Post a Job",
                                icon: "plus.circle.fill",
                                color: .green
                            ) {
                                showingCreateJob = true
                            }
                        }
                    }
                    .padding(.horizontal)
                    
                    // Stats Section
                    VStack(alignment: .leading, spacing: 12) {
                        Text("Quick Stats")
                            .font(.headline)
                            .padding(.horizontal)
                        
                        HStack(spacing: 16) {
                            if authService.currentUser?.role == "consumer" {
                                StatCard(
                                    title: "My Active Jobs",
                                    value: "\(activeJobsCount)",
                                    icon: "briefcase",
                                    color: .blue
                                )

                                StatCard(
                                    title: "Completed Jobs",
                                    value: "\(completedJobsCount)",
                                    icon: "checkmark.circle",
                                    color: .green
                                )
                            } else {
                                StatCard(
                                    title: "Jobs Accepted",
                                    value: "\(activeJobsCount)",
                                    icon: "briefcase",
                                    color: .blue
                                )

                                StatCard(
                                    title: "Jobs Completed",
                                    value: "\(completedJobsCount)",
                                    icon: "checkmark.circle",
                                    color: .green
                                )
                            }
                        }
                        .padding(.horizontal)
                    }
                    
                    // Recent Activity
                    VStack(alignment: .leading, spacing: 12) {
                        Text("Recent Activity")
                            .font(.headline)
                            .padding(.horizontal)
                        
                        VStack(spacing: 8) {
                            ActivityItemView(
                                title: "Welcome to GigCo!",
                                subtitle: "Complete your profile to get started",
                                time: "Now"
                            )
                        }
                    }
                    
                    Spacer()
                }
                .padding(.top)
            }
            .navigationTitle("Dashboard")
            .sheet(isPresented: $showingJobList) {
                JobListView()
                    .environmentObject(authService)
            }
            .sheet(isPresented: $showingCreateJob) {
                CreateJobView()
                    .environmentObject(authService)
            }
            .task {
                await loadDashboardData()
            }
            .onAppear {
                // Refresh stats when returning to home tab
                Task {
                    await loadDashboardData()
                }
            }
        }
    }
    
    private func loadDashboardData() async {
        // Perform API health check
        do {
            let _ = try await apiService.healthCheck()
            await MainActor.run {
                apiHealthy = true
                healthCheckMessage = "API Connected"
            }
        } catch {
            await MainActor.run {
                apiHealthy = false
                healthCheckMessage = "API Disconnected"
            }
        }
        
        // Load job data for current user
        if let currentUser = authService.currentUser,
           let userID = currentUser.id {
            do {
                print("🔵 DashboardView - Loading jobs for user: \(currentUser.name), ID: \(userID), role: \(currentUser.role)")
                try await jobService.getMyJobs(for: userID, role: currentUser.role)

                await MainActor.run {
                    // Calculate active jobs (posted, accepted, in_progress)
                    let activeStatuses = ["posted", "accepted", "in_progress"]
                    activeJobsCount = jobService.myJobs.filter { job in
                        guard let status = job.status else { return false }
                        return activeStatuses.contains(status.lowercased())
                    }.count

                    // Calculate completed jobs
                    completedJobsCount = jobService.myJobs.filter { job in
                        guard let status = job.status else { return false }
                        return status.lowercased() == "completed"
                    }.count

                    print("🟢 DashboardView - Stats updated: Active: \(activeJobsCount), Completed: \(completedJobsCount)")
                }
            } catch {
                print("🔴 DashboardView - Failed to load job stats: \(error)")
                await MainActor.run {
                    activeJobsCount = 0
                    completedJobsCount = 0
                }
            }
        } else {
            print("🔴 DashboardView - No current user found for loading job stats")
            await MainActor.run {
                activeJobsCount = 0
                completedJobsCount = 0
            }
        }
    }
}

struct StatCard: View {
    let title: String
    let value: String
    let icon: String
    let color: Color
    
    var body: some View {
        VStack(spacing: 8) {
            Image(systemName: icon)
                .font(.title2)
                .foregroundColor(color)
            
            Text(value)
                .font(.title2)
                .fontWeight(.bold)
                .foregroundColor(.primary)
            
            Text(title)
                .font(.caption)
                .foregroundColor(.secondary)
        }
        .frame(maxWidth: .infinity)
        .padding()
        .background(Color(.systemGray6))
        .cornerRadius(12)
    }
}

struct JobsTabView: View {
    var body: some View {
        NavigationView {
            VStack {
                Text("Jobs will be displayed here")
                    .font(.title2)
                    .foregroundColor(.gray)
                    .padding()
                
                Spacer()
            }
            .navigationTitle("Jobs")
        }
    }
}

struct ProfileTabView: View {
    @EnvironmentObject var authService: AuthService
    
    var body: some View {
        NavigationView {
            VStack(spacing: 20) {
                // Profile Header
                VStack(spacing: 12) {
                    Circle()
                        .fill(Color.blue.opacity(0.2))
                        .frame(width: 80, height: 80)
                        .overlay(
                            Text(userInitials)
                                .font(.title)
                                .fontWeight(.bold)
                                .foregroundColor(.blue)
                        )
                    
                    if let user = authService.currentUser {
                        Text(user.name)
                            .font(.title2)
                            .fontWeight(.semibold)
                        
                        Text(user.email)
                            .font(.subheadline)
                            .foregroundColor(.gray)
                        
                        Text(user.role.capitalized)
                            .font(.caption)
                            .padding(.horizontal, 12)
                            .padding(.vertical, 4)
                            .background(Color.blue.opacity(0.1))
                            .foregroundColor(.blue)
                            .cornerRadius(12)
                    }
                }
                .padding(.top, 20)
                
                // Profile Options
                VStack(spacing: 0) {
                    ProfileOptionRow(icon: "person.fill", title: "Edit Profile") { }
                    ProfileOptionRow(icon: "bell.fill", title: "Notifications") { }
                    ProfileOptionRow(icon: "gear", title: "Settings") { }
                    ProfileOptionRow(icon: "questionmark.circle", title: "Help & Support") { }
                }
                .background(Color.white)
                .cornerRadius(12)
                .shadow(radius: 1)
                .padding(.horizontal)
                
                Spacer()
                
                // Logout Button
                Button(action: {
                    authService.logout()
                }) {
                    Text("Logout")
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color.red)
                        .foregroundColor(.white)
                        .cornerRadius(8)
                }
                .padding(.horizontal)
                .padding(.bottom, 20)
            }
            .navigationTitle("Profile")
            .background(Color(.systemGroupedBackground))
        }
    }
    
    private var userInitials: String {
        guard let user = authService.currentUser else { return "U" }
        let nameComponents = user.name.components(separatedBy: " ")
        let firstInitial = nameComponents.first?.prefix(1).uppercased() ?? "U"
        let lastInitial = nameComponents.count > 1 ? nameComponents.last?.prefix(1).uppercased() ?? "" : ""
        return "\(firstInitial)\(lastInitial)"
    }
}

struct QuickActionCard: View {
    let title: String
    let icon: String
    let color: Color
    let action: () -> Void
    
    var body: some View {
        Button(action: action) {
            VStack(spacing: 12) {
                Image(systemName: icon)
                    .font(.title2)
                    .foregroundColor(color)
                    .frame(width: 32, height: 32)
                
                Text(title)
                    .font(.headline)
                    .multilineTextAlignment(.center)
                    .foregroundColor(.primary)
                    .lineLimit(2)
            }
            .frame(minWidth: 140, maxWidth: .infinity)
            .frame(height: 120)
            .padding()
            .background(Color.white)
            .cornerRadius(12)
            .shadow(color: .gray.opacity(0.3), radius: 2, x: 0, y: 1)
        }
        .buttonStyle(PlainButtonStyle())
    }
}

struct ActivityItemView: View {
    let title: String
    let subtitle: String
    let time: String
    
    var body: some View {
        HStack {
            Circle()
                .fill(Color.blue)
                .frame(width: 8, height: 8)
            
            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                    .font(.subheadline)
                    .fontWeight(.medium)
                
                Text(subtitle)
                    .font(.caption)
                    .foregroundColor(.gray)
            }
            
            Spacer()
            
            Text(time)
                .font(.caption)
                .foregroundColor(.gray)
        }
        .padding(.horizontal)
        .padding(.vertical, 8)
        .background(Color.white)
        .cornerRadius(8)
        .shadow(radius: 1)
    }
}

struct ProfileOptionRow: View {
    let icon: String
    let title: String
    let action: () -> Void
    
    var body: some View {
        Button(action: action) {
            HStack(spacing: 16) {
                Image(systemName: icon)
                    .foregroundColor(.blue)
                    .frame(width: 24)
                
                Text(title)
                    .foregroundColor(.primary)
                
                Spacer()
                
                Image(systemName: "chevron.right")
                    .foregroundColor(.gray)
                    .font(.caption)
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 12)
        }
        .buttonStyle(PlainButtonStyle())
    }
}