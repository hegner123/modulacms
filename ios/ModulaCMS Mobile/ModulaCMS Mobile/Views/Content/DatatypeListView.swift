import SwiftUI
import ModulaCMS

struct DatatypeListView: View {
    @State private var viewModel = DatatypeListViewModel()

    var body: some View {
        LoadStateView(state: viewModel.loadState, emptyMessage: "No content types found") { datatypes in
            List(datatypes, id: \.datatypeID.rawValue) { datatype in
                NavigationLink(value: datatype) {
                    Label(datatype.label, systemImage: "folder")
                }
            }
            .navigationDestination(for: Datatype.self) { datatype in
                ContentListView(datatypeID: datatype.datatypeID, datatypeLabel: datatype.label)
            }
        }
        .navigationTitle("Content")
        .task { await viewModel.load() }
        .refreshable { await viewModel.load() }
    }
}

extension Datatype: @retroactive Hashable {
    public func hash(into hasher: inout Hasher) {
        hasher.combine(datatypeID.rawValue)
    }

    public static func == (lhs: Datatype, rhs: Datatype) -> Bool {
        lhs.datatypeID.rawValue == rhs.datatypeID.rawValue
    }
}
