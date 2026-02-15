import SwiftUI
import ModulaCMS

struct RouteListView: View {
    @State private var viewModel = RouteListViewModel()
    @State private var showingCreate = false
    @State private var resolvingRoute: Route?
    @State private var resolvedContentID: ContentID?
    @State private var navigationPath = NavigationPath()

    var body: some View {
        LoadStateView(state: viewModel.loadState, emptyMessage: "No routes found") { routes in
            List {
                ForEach(routes, id: \.routeID.rawValue) { route in
                    Button {
                        resolveAndNavigate(route)
                    } label: {
                        HStack {
                            VStack(alignment: .leading, spacing: 4) {
                                Text(route.title)
                                    .font(.headline)
                                Text(route.slug.rawValue)
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }
                            Spacer()
                            Text(route.status == 1 ? "Active" : "Inactive")
                                .font(.caption)
                                .padding(.horizontal, 8)
                                .padding(.vertical, 4)
                                .background(route.status == 1 ? Color.green.opacity(0.2) : Color.gray.opacity(0.2))
                                .clipShape(Capsule())
                        }
                    }
                    .foregroundStyle(.primary)
                }
                .onDelete { offsets in
                    guard case .loaded(let routes) = viewModel.loadState else { return }
                    Task {
                        for index in offsets {
                            try? await viewModel.deleteRoute(at: routes[index].routeID)
                        }
                    }
                }
            }
            .navigationDestination(for: ContentID.self) { id in
                ContentEditorView(contentDataID: id)
            }
        }
        .navigationTitle("Routes")
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
                RouteCreateView()
            }
        }
        .task { await viewModel.load() }
        .refreshable { await viewModel.load() }
    }

    private func resolveAndNavigate(_ route: Route) {
        Task {
            if let contentID = await viewModel.resolveContentDataID(for: route) {
                resolvedContentID = contentID
            }
        }
    }
}
