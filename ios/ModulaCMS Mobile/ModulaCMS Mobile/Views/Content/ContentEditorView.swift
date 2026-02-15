import SwiftUI
import ModulaCMS

struct ContentEditorView: View {
    @State private var viewModel: ContentEditorViewModel
    @State private var showDeleteConfirmation = false
    @Environment(\.dismiss) private var dismiss

    init(contentDataID: ContentID) {
        _viewModel = State(initialValue: ContentEditorViewModel(contentDataID: contentDataID))
    }

    var body: some View {
        LoadStateView(state: viewModel.loadState, emptyMessage: "Content not found") { _ in
            Form {
                if let error = viewModel.saveError {
                    ErrorBanner(message: error)
                }

                routeSection
                contentSection
                fieldsSection
            }
        }
        .navigationTitle(viewModel.route?.title ?? "Content")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                if viewModel.isEditing {
                    Button("Save") {
                        Task { await viewModel.save() }
                    }
                    .disabled(viewModel.isSaving)
                } else {
                    Button("Edit") {
                        viewModel.isEditing = true
                    }
                }
            }
            ToolbarItem(placement: .cancellationAction) {
                if viewModel.isEditing {
                    Button("Cancel") {
                        viewModel.isEditing = false
                        viewModel.saveError = nil
                    }
                }
            }
            ToolbarItem(placement: .destructiveAction) {
                if !viewModel.isEditing {
                    ConfirmDeleteButton(
                        title: "Delete Content",
                        message: "This will permanently delete this content and its fields."
                    ) {
                        Task { await viewModel.delete() }
                    }
                }
            }
        }
        .task { await viewModel.load() }
        .refreshable { await viewModel.load() }
        .onChange(of: viewModel.isDeleted) { _, deleted in
            if deleted { dismiss() }
        }
    }

    @ViewBuilder
    private var routeSection: some View {
        if let route = viewModel.route {
            Section("Route") {
                if viewModel.isEditing {
                    TextField("Title", text: $viewModel.editTitle)
                    TextField("Slug", text: $viewModel.editSlug)
                    Picker("Status", selection: $viewModel.editRouteStatus) {
                        Text("Active").tag(Int64(1))
                        Text("Inactive").tag(Int64(0))
                    }
                } else {
                    LabeledContent("Title", value: route.title)
                    LabeledContent("Slug", value: route.slug.rawValue)
                    LabeledContent("Status", value: route.status == 1 ? "Active" : "Inactive")
                }
            }
        }
    }

    @ViewBuilder
    private var contentSection: some View {
        if let content = viewModel.contentData {
            Section("Content") {
                if viewModel.isEditing {
                    Picker("Status", selection: $viewModel.editContentStatus) {
                        Text("Draft").tag(ContentStatus.draft)
                        Text("Published").tag(ContentStatus.published)
                        Text("Archived").tag(ContentStatus.archived)
                        Text("Pending").tag(ContentStatus.pending)
                    }
                } else {
                    HStack {
                        Text("Status")
                        Spacer()
                        StatusBadge(status: content.status)
                    }
                }
                if !viewModel.datatypeLabel.isEmpty {
                    LabeledContent("Type", value: viewModel.datatypeLabel)
                }
                LabeledContent("Modified", value: viewModel.contentData?.dateModified.rawValue ?? "")
            }
        }
    }

    @ViewBuilder
    private var fieldsSection: some View {
        if !viewModel.fieldPairs.isEmpty {
            Section("Fields") {
                if viewModel.isEditing {
                    ContentFieldEditView(
                        fieldPairs: viewModel.fieldPairs,
                        values: $viewModel.editFieldValues
                    )
                } else {
                    ContentFieldsSection(fieldPairs: viewModel.fieldPairs)
                }
            }
        }
    }
}
