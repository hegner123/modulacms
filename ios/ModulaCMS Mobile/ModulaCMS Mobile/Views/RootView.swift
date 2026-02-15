import SwiftUI

enum AppTab: Hashable {
    case content
    case routes
    case settings
}

struct RootView: View {
    @State private var selectedTab: AppTab = .content
    @State private var showWelcome = false

    var body: some View {
        TabView(selection: $selectedTab) {
            Tab("Content", systemImage: "list.bullet", value: .content) {
                NavigationStack {
                    DatatypeListView()
                }
            }
            Tab("Routes", systemImage: "globe", value: .routes) {
                NavigationStack {
                    RouteListView()
                }
            }
            Tab("Settings", systemImage: "gear", value: .settings) {
                NavigationStack {
                    SettingsView()
                }
            }
        }
        .onAppear {
            if AppConfig.needsSetup {
                showWelcome = true
            }
        }
        .sheet(isPresented: $showWelcome) {
            WelcomeSheet {
                showWelcome = false
                selectedTab = .settings
            }
        }
    }
}

private struct WelcomeSheet: View {
    var onConfigure: () -> Void
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        VStack(spacing: 24) {
            Spacer()

            Image(systemName: "server.rack")
                .font(.system(size: 56))
                .foregroundStyle(.tint)

            Text("Get Started")
                .font(.title)
                .fontWeight(.bold)

            Text("Connect to your ModulaCMS backend to manage content, routes, and more.")
                .multilineTextAlignment(.center)
                .foregroundStyle(.secondary)
                .padding(.horizontal, 32)

            Spacer()

            Button {
                onConfigure()
            } label: {
                Text("Configure Server")
                    .fontWeight(.semibold)
                    .frame(maxWidth: .infinity)
            }
            .buttonStyle(.borderedProminent)
            .controlSize(.large)

            Button("Skip for Now") {
                dismiss()
            }
            .foregroundStyle(.secondary)

            Spacer()
                .frame(height: 16)
        }
        .padding(.horizontal, 24)
        .interactiveDismissDisabled(false)
    }
}
