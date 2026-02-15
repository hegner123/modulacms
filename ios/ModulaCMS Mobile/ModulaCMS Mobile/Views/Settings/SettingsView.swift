import SwiftUI

struct SettingsView: View {
    @State private var viewModel = SettingsViewModel()

    var body: some View {
        Form {
            Section("Server") {
                TextField("Base URL", text: $viewModel.baseURL)
                    .keyboardType(.URL)
                    .textInputAutocapitalization(.never)
                    .autocorrectionDisabled()
                HStack {
                    TextField("API Key", text: $viewModel.apiKey)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled()
                    if !viewModel.apiKey.isEmpty {
                        Button {
                            viewModel.apiKey = ""
                        } label: {
                            Image(systemName: "xmark.circle.fill")
                                .foregroundStyle(.secondary)
                        }
                        .buttonStyle(.plain)
                    }
                }
            }

            Section("Dev") {
                TextField("Author ID", text: $viewModel.authorID)
                    .textInputAutocapitalization(.never)
                    .autocorrectionDisabled()
                    .font(.system(.body, design: .monospaced))
            }

            Section("Connection") {
                Button("Test Connection") {
                    Task { await viewModel.testConnection() }
                }
                .disabled(viewModel.baseURL.isEmpty)

                switch viewModel.testState {
                case .idle:
                    EmptyView()
                case .loading:
                    HStack {
                        ProgressView()
                        Text("Testing...")
                            .foregroundStyle(.secondary)
                    }
                case .loaded(let message):
                    Label(message, systemImage: "checkmark.circle.fill")
                        .foregroundStyle(.green)
                case .error(let message):
                    Label(message, systemImage: "xmark.circle.fill")
                        .foregroundStyle(.red)
                }
            }
        }
        .navigationTitle("Settings")
        .toolbar {
            ToolbarItem(placement: .confirmationAction) {
                Button("Save") {
                    viewModel.save()
                }
                .disabled(!viewModel.hasChanges)
            }
        }
        .overlay {
            if viewModel.showSaveConfirmation {
                savedToast
            }
        }
    }

    private var savedToast: some View {
        Text("Settings Saved")
            .font(.subheadline)
            .fontWeight(.medium)
            .padding(.horizontal, 16)
            .padding(.vertical, 10)
            .background(.thinMaterial, in: Capsule())
            .frame(maxHeight: .infinity, alignment: .bottom)
            .padding(.bottom, 32)
            .transition(.move(edge: .bottom).combined(with: .opacity))
            .onAppear {
                DispatchQueue.main.asyncAfter(deadline: .now() + 1.5) {
                    withAnimation { viewModel.showSaveConfirmation = false }
                }
            }
    }
}
