import SwiftUI

struct RouteCreateView: View {
    @State private var viewModel = RouteCreateViewModel()
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        Form {
            if let error = viewModel.saveError {
                ErrorBanner(message: error)
            }

            Section("Route") {
                TextField("Title", text: $viewModel.title)
                TextField("Slug", text: $viewModel.slug, prompt: Text(viewModel.slugFromTitle))
            }

            Section("Status") {
                Picker("Status", selection: $viewModel.status) {
                    Text("Active").tag(Int64(1))
                    Text("Inactive").tag(Int64(0))
                }
                .pickerStyle(.menu)
            }
        }
        .navigationTitle("New Route")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button("Cancel") { dismiss() }
            }
            ToolbarItem(placement: .confirmationAction) {
                Button("Create") {
                    Task { await viewModel.create() }
                }
                .disabled(viewModel.title.isEmpty || viewModel.isSaving)
            }
        }
        .onChange(of: viewModel.didCreate) { _, created in
            if created { dismiss() }
        }
    }
}
