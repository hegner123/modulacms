import Foundation
import ModulaCMS

@Observable
final class DatatypeListViewModel {
    var loadState: LoadState<[Datatype]> = .idle

    func load() async {
        loadState = .loading
        do {
            let all = try await CMSClient.shared.datatypes.list()
            let roots = all.filter { $0.type.lowercased() == "root" }
            loadState = .loaded(roots)
        } catch {
            loadState = .error(error.localizedDescription)
        }
    }
}
