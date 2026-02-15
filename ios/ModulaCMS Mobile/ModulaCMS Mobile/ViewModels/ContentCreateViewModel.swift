import Foundation
import ModulaCMS

@Observable
final class ContentCreateViewModel {
    let datatypeID: DatatypeID
    var loadState: LoadState<[FieldPair]> = .idle
    var isSaving = false
    var saveError: String?
    var didCreate = false

    var title = ""
    var slug = ""
    var status: ContentStatus = .draft
    var fieldValues: [String: String] = [:]

    init(datatypeID: DatatypeID) {
        self.datatypeID = datatypeID
    }

    func loadFields() async {
        loadState = .loading
        do {
            async let fieldsResult = CMSClient.shared.fields.list()
            async let dtFieldsResult = CMSClient.shared.datatypeFields.list()
            let (allFields, allDtFields) = try await (fieldsResult, dtFieldsResult)

            let fieldMap = Dictionary(
                allFields.map { ($0.fieldID.rawValue, $0) },
                uniquingKeysWith: { first, _ in first }
            )

            let pairs = allDtFields
                .filter { $0.datatypeID.rawValue == datatypeID.rawValue }
                .sorted { $0.sortOrder < $1.sortOrder }
                .compactMap { dtf -> FieldPair? in
                    guard let field = fieldMap[dtf.fieldID.rawValue] else { return nil }
                    return FieldPair(field: field, contentField: nil, sortOrder: dtf.sortOrder)
                }

            for pair in pairs {
                fieldValues[pair.field.fieldID.rawValue] = pair.field.type == .boolean ? "false" : ""
            }

            loadState = .loaded(pairs)
        } catch {
            loadState = .error(error.localizedDescription)
        }
    }

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

            let authorID = UserID(AppConfig.authorID)

            let route = try await CMSClient.shared.routes.create(params: CreateRouteParams(
                slug: Slug(effectiveSlug),
                title: title,
                status: 1,
                authorID: authorID
            ))

            let content = try await CMSClient.shared.contentData.create(params: CreateContentDataParams(
                routeID: route.routeID,
                datatypeID: datatypeID,
                authorID: authorID,
                status: status
            ))

            guard case .loaded(let pairs) = loadState else { return }
            for pair in pairs {
                let val = fieldValues[pair.field.fieldID.rawValue] ?? ""
                if !val.isEmpty {
                    _ = try await CMSClient.shared.contentFields.create(params: CreateContentFieldParams(
                        contentDataID: content.contentDataID,
                        fieldID: pair.field.fieldID,
                        fieldValue: val,
                        authorID: authorID
                    ))
                }
            }

            isSaving = false
            didCreate = true
        } catch {
            isSaving = false
            saveError = error.localizedDescription
        }
    }
}
