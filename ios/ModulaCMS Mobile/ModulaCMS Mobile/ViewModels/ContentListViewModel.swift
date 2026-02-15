import Foundation
import ModulaCMS

@Observable
final class ContentListViewModel {
    let datatypeID: DatatypeID
    let datatypeLabel: String
    var loadState: LoadState<[ContentItem]> = .idle

    init(datatypeID: DatatypeID, datatypeLabel: String) {
        self.datatypeID = datatypeID
        self.datatypeLabel = datatypeLabel
    }

    func load() async {
        loadState = .loading
        do {
            async let contentList = CMSClient.shared.contentData.list()
            async let routeList = CMSClient.shared.routes.list()
            let (allContent, allRoutes) = try await (contentList, routeList)

            let routeMap = Dictionary(
                allRoutes.map { ($0.routeID.rawValue, $0) },
                uniquingKeysWith: { first, _ in first }
            )

            let items = allContent
                .filter { $0.datatypeID?.rawValue == datatypeID.rawValue }
                .map { content in
                    let route = content.routeID.flatMap { routeMap[$0.rawValue] }
                    return ContentItem(
                        contentData: content,
                        route: route,
                        datatypeLabel: datatypeLabel
                    )
                }
                .sorted { $0.dateModified.rawValue > $1.dateModified.rawValue }

            loadState = .loaded(items)
        } catch {
            loadState = .error(error.localizedDescription)
        }
    }

    func deleteContent(at id: ContentID) async throws {
        try await CMSClient.shared.contentData.delete(id: id)
        await load()
    }
}
