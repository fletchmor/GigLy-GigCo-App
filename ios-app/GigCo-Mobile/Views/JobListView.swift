//
//  JobListView.swift
//  GigCo-Mobile
//
//  Created by Claude on 9/9/25.
//

import SwiftUI

struct JobListView: View {
    @StateObject private var jobService = JobService()
    @EnvironmentObject var authService: AuthService
    @State private var showingCreateJob = false
    @State private var showError = false
    @State private var errorMessage = ""
    @State private var selectedTab = 0
    
    var body: some View {
        NavigationView {
            VStack(spacing: 0) {
                // Tab Picker
                Picker("Job Type", selection: $selectedTab) {
                    Text("All Jobs").tag(0)
                    if authService.currentUser?.role == "gig_worker" {
                        Text("Available").tag(1)
                    }
                    Text("My Jobs").tag(2)
                }
                .pickerStyle(SegmentedPickerStyle())
                .padding()
                
                // Content based on selected tab
                if jobService.isLoading {
                    Spacer()
                    ProgressView("Loading jobs...")
                    Spacer()
                } else if currentJobs.isEmpty {
                    VStack(spacing: 16) {
                        Spacer()
                        Image(systemName: "briefcase.fill")
                            .font(.system(size: 50))
                            .foregroundColor(.gray)
                        Text("No jobs found")
                            .font(.title2)
                            .foregroundColor(.gray)
                        if selectedTab == 2 && authService.currentUser?.role == "consumer" {
                            Text("Create your first job posting!")
                                .font(.subheadline)
                                .foregroundColor(.secondary)
                            Button("Post a Job") {
                                showingCreateJob = true
                            }
                            .buttonStyle(.borderedProminent)
                        } else {
                            Text("Pull to refresh")
                                .font(.subheadline)
                                .foregroundColor(.secondary)
                        }
                        Spacer()
                    }
                    .frame(maxWidth: .infinity)
                } else {
                    List {
                        ForEach(currentJobs) { job in
                            JobRowView(job: job, jobService: jobService)
                                .environmentObject(authService)
                        }
                    }
                    .refreshable {
                        await loadJobs()
                    }
                }
            }
            .navigationTitle("Jobs")
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    if authService.currentUser?.role == "consumer" {
                        Button("Post Job") {
                            showingCreateJob = true
                        }
                    }
                }
            }
            .sheet(isPresented: $showingCreateJob) {
                CreateJobView()
                    .environmentObject(authService)
            }
            .alert("Error", isPresented: $showError) {
                Button("OK") { }
            } message: {
                Text(errorMessage)
            }
            .task {
                await loadJobs()
            }
            .onChange(of: selectedTab) {
                Task {
                    await loadJobs()
                }
            }
        }
    }
    
    private var currentJobs: [Job] {
        let jobs: [Job]
        switch selectedTab {
        case 0:
            jobs = jobService.jobs
        case 1:
            jobs = jobService.availableJobs
        case 2:
            jobs = jobService.myJobs
        default:
            jobs = jobService.jobs
        }

        print("🔵 JobListView.currentJobs - selectedTab: \(selectedTab), jobs count: \(jobs.count)")
        print("🔵 JobListView.currentJobs - isLoading: \(jobService.isLoading)")
        return jobs
    }
    
    private func loadJobs() async {
        do {
            switch selectedTab {
            case 0:
                try await jobService.getAllJobs()
            case 1:
                if authService.currentUser?.role == "gig_worker" {
                    try await jobService.getAvailableJobs()
                }
            case 2:
                if let currentUser = authService.currentUser,
                   let userID = currentUser.id {
                    print("🔵 JobListView - Loading My Jobs for user: \(currentUser.name), ID: \(userID), role: \(currentUser.role)")
                    try await jobService.getMyJobs(for: userID, role: currentUser.role)
                } else {
                    print("🔴 JobListView - No current user or user ID found")
                    print("🔴 JobListView - currentUser: \(authService.currentUser?.name ?? "nil")")
                    print("🔴 JobListView - currentUser.id: \(authService.currentUser?.id?.description ?? "nil")")
                    try await jobService.getMyJobs()
                }
            default:
                try await jobService.getAllJobs()
            }
        } catch {
            errorMessage = error.localizedDescription
            showError = true
        }
    }
}