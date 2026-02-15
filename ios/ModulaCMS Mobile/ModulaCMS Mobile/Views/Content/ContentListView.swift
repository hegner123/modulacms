import SwiftUI
import ModulaCMS

struct ContentListView: View {
    @State private var viewModel: ContentListViewModel
    @State private var showingCreate = false

    init(datatypeID: DatatypeID, datatypeLabel: String) {
        _viewModel = State(initialValue: ContentListViewModel(
            datatypeID: datatypeID,
            datatypeLabel: datatypeLabel
        ))
    }

    var body: some View {
        LoadStateView(state: viewModel.loadState, emptyMessage: "No content found") { items in
            List {
                ForEach(items) { item in
                    NavigationLink(value: item.contentData.contentDataID) {
                        HStack {
                            VStack(alignment: .leading, spacing: 4) {
                                Text(item.title)
                                    .font(.headline)
                                if !item.slug.isEmpty {
                                    Text(item.slug)
                                        .font(.caption)
                                        .foregroundStyle(.secondary)
                                }
                            }
                            Spacer()
                            StatusBadge(status: item.status)
                        }
                    }
                }
                .onDelete { offsets in
                    guard case .loaded(let items) = viewModel.loadState else { return }
                    Task {
                        for index in offsets {
                            try? await viewModel.deleteContent(at: items[index].contentData.contentDataID)
                        }
                    }
                }
            }
        }
        .navigationTitle(viewModel.datatypeLabel)
        .navigationDestination(for: ContentID.self) { id in
            ContentEditorView(contentDataID: id)
        }
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button {
                    showingCreate = true
                } label: {
                    Label("Add", systemImage: "plus")
                }
            }
        }
        .sheet(isPresented: $showingCreate) {
            Task { await viewModel.load() }
        } content: {
            NavigationStack {
                ContentCreateView(
                    datatypeID: viewModel.datatypeID,
                    datatypeLabel: viewModel.datatypeLabel
                )
            }
        }
        .task { await viewModel.load() }
        .refreshable { await viewModel.load() }
    }
}
