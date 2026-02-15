import SwiftUI
import ModulaCMS

struct ContentCreateView: View {
    @State private var viewModel: ContentCreateViewModel
    @Environment(\.dismiss) private var dismiss
    let datatypeLabel: String

    init(datatypeID: DatatypeID, datatypeLabel: String) {
        _viewModel = State(initialValue: ContentCreateViewModel(datatypeID: datatypeID))
        self.datatypeLabel = datatypeLabel
    }

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
                Picker("Content Status", selection: $viewModel.status) {
                    Text("Draft").tag(ContentStatus.draft)
                    Text("Published").tag(ContentStatus.published)
                    Text("Archived").tag(ContentStatus.archived)
                    Text("Pending").tag(ContentStatus.pending)
                }
                .pickerStyle(.menu)
            }

            fieldsSection
        }
        .navigationTitle("New \(datatypeLabel)")
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
        .task { await viewModel.loadFields() }
        .onChange(of: viewModel.didCreate) { _, created in
            if created { dismiss() }
        }
    }

    @ViewBuilder
    private var fieldsSection: some View {
        if case .loaded(let pairs) = viewModel.loadState, !pairs.isEmpty {
            Section("Fields") {
                ForEach(pairs) { pair in
                    FieldRenderer(
                        field: pair.field,
                        value: binding(for: pair.field.fieldID.rawValue),
                        readOnly: false
                    )
                }
            }
        }
    }

    private func binding(for key: String) -> Binding<String> {
        Binding(
            get: { viewModel.fieldValues[key] ?? "" },
            set: { viewModel.fieldValues[key] = $0 }
        )
    }
}
