import Foundation
import ModulaCMS

@Observable
final class RouteCreateViewModel {
    var title = ""
    var slug = ""
    var status: Int64 = 1
    var isSaving = false
    var saveError: String?
    var didCreate = false

    var slugFromTitle: String {
        title.lowercased()
            .replacingOccurrences(of: " ", with: "-")
            .filter { $0.isLetter || $0.isNumber || $0 == "-" }
    }

    func create() async {
        isSaving = true
        saveError = nil
        do {
            let rawSlug = slug.isEmpty ? slugFromTitle : slug
            let effectiveSlug = rawSlug.hasPrefix("/") ? rawSlug : "/" + rawSlug
            _ = try await CMSClient.shared.routes.create(params: CreateRouteParams(
                slug: Slug(effectiveSlug),
                title: title,
                status: status,
                authorID: UserID(AppConfig.authorID)
            ))
            isSaving = false
            didCreate = true
        } catch {
            isSaving = false
            saveError = error.localizedDescription
        }
    }
}
