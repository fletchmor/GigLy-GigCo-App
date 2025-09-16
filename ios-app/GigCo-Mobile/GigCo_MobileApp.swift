//
//  GigCo_MobileApp.swift
//  GigCo-Mobile
//
//  Created by Fletcher Morris on 9/8/25.
//

import SwiftUI
import SwiftData

@main
struct GigCo_MobileApp: App {
    @StateObject private var authService = AuthService()
    @StateObject private var userService = UserService.shared
    @StateObject private var apiService = APIService.shared
    
    var sharedModelContainer: ModelContainer = {
        let schema = Schema([
            Item.self,
        ])
        let modelConfiguration = ModelConfiguration(schema: schema, isStoredInMemoryOnly: false)

        do {
            return try ModelContainer(for: schema, configurations: [modelConfiguration])
        } catch {
            fatalError("Could not create ModelContainer: \(error)")
        }
    }()

    var body: some Scene {
        WindowGroup {
            RootView()
                .environmentObject(authService)
                .environmentObject(userService)
                .environmentObject(apiService)
        }
        .modelContainer(sharedModelContainer)
    }
}

struct RootView: View {
    @EnvironmentObject var authService: AuthService
    
    var body: some View {
        Group {
            if authService.isAuthenticated {
                DashboardView()
            } else {
                LoginView()
            }
        }
    }
}
