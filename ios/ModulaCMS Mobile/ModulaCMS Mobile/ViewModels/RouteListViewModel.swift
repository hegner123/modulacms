import Foundation
import ModulaCMS

@Observable
final class RouteListViewModel {
    var loadState: LoadState<[Route]> = .idle

    func load() async {
        loadState = .loading
        do {
            let routes = try await CMSClient.shared.routes.list()
            let sorted = routes.sorted { $0.dateModified.rawValue > $1.dateModified.rawValue }
            loadState = .loaded(sorted)
        } catch {
            loadState = .error(error.localizedDescription)
        }
    }

    func resolveContentDataID(for route: Route) async -> ContentID? {
        do {
            let allContent = try await CMSClient.shared.contentData.list()
            let allDatatypes = try await CMSClient.shared.datatypes.list()
            let rootDatatypeIDs = Set(
                allDatatypes.filter { $0.type.lowercased() == "root" }.map { $0.datatypeID.rawValue }
            )
            let match = allContent.first { cd in
                cd.routeID?.rawValue == route.routeID.rawValue &&
                cd.datatypeID.map({ rootDatatypeIDs.contains($0.rawValue) }) == true
            }
            return match?.contentDataID
        } catch {
            return nil
        }
    }

    func deleteRoute(at id: RouteID) async throws {
        try await CMSClient.shared.routes.delete(id: id)
        await load()
    }
}
