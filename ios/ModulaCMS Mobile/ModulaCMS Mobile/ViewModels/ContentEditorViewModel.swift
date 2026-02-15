import Foundation
import ModulaCMS

struct FieldPair: Identifiable {
    let field: Field
    let contentField: ContentField?
    let sortOrder: Int64

    var id: String { field.fieldID.rawValue }
    var value: String { contentField?.fieldValue ?? "" }
}

@Observable
final class ContentEditorViewModel {
    let contentDataID: ContentID
    var loadState: LoadState<Void> = .idle
    var isEditing = false
    var isSaving = false
    var saveError: String?
    var isDeleted = false

    var contentData: ContentData?
    var route: Route?
    var datatypeLabel = ""
    var fieldPairs: [FieldPair] = []

    // Edit state
    var editTitle = ""
    var editSlug = ""
    var editRouteStatus: Int64 = 1
    var editContentStatus: ContentStatus = .draft
    var editFieldValues: [String: String] = [:]

    init(contentDataID: ContentID) {
        self.contentDataID = contentDataID
    }

    func load() async {
        loadState = .loading
        do {
            async let contentResult = CMSClient.shared.contentData.get(id: contentDataID)
            async let fieldsResult = CMSClient.shared.fields.list()
            async let contentFieldsResult = CMSClient.shared.contentFields.list()
            async let datatypeFieldsResult = CMSClient.shared.datatypeFields.list()
            async let routesResult = CMSClient.shared.routes.list()
            async let datatypesResult = CMSClient.shared.datatypes.list()

            let (content, allFields, allContentFields, allDatatypeFields, allRoutes, allDatatypes) =
                try await (contentResult, fieldsResult, contentFieldsResult, datatypeFieldsResult, routesResult, datatypesResult)

            contentData = content
            route = content.routeID.flatMap { rid in allRoutes.first { $0.routeID.rawValue == rid.rawValue } }

            if let dtID = content.datatypeID {
                datatypeLabel = allDatatypes.first { $0.datatypeID.rawValue == dtID.rawValue }?.label ?? ""
            }

            let fieldMap = Dictionary(
                allFields.map { ($0.fieldID.rawValue, $0) },
                uniquingKeysWith: { first, _ in first }
            )
            let contentFieldMap = Dictionary(
                allContentFields
                    .filter { $0.contentDataID?.rawValue == contentDataID.rawValue }
                    .map { ($0.fieldID?.rawValue ?? "", $0) },
                uniquingKeysWith: { first, _ in first }
            )

            let dtFields = allDatatypeFields
                .filter { $0.datatypeID.rawValue == content.datatypeID?.rawValue }
                .sorted { $0.sortOrder < $1.sortOrder }

            fieldPairs = dtFields.compactMap { dtf in
                guard let field = fieldMap[dtf.fieldID.rawValue] else { return nil }
                let cf = contentFieldMap[dtf.fieldID.rawValue]
                return FieldPair(field: field, contentField: cf, sortOrder: dtf.sortOrder)
            }

            // Initialize edit state from loaded data
            editTitle = route?.title ?? ""
            editSlug = route?.slug.rawValue ?? ""
            editRouteStatus = route?.status ?? 1
            editContentStatus = content.status
            editFieldValues = [:]
            for pair in fieldPairs {
                editFieldValues[pair.field.fieldID.rawValue] = pair.value
            }

            loadState = .loaded(())
        } catch {
            loadState = .error(error.localizedDescription)
        }
    }

    func delete() async {
        do {
            try await CMSClient.shared.contentData.delete(id: contentDataID)
            isDeleted = true
        } catch {
            saveError = error.localizedDescription
        }
    }

    private var userID: UserID {
        UserID(AppConfig.authorID)
    }

    func save() async {
        guard let content = contentData else { return }
        isSaving = true
        saveError = nil
        do {
            // Update route if changed
            if let currentRoute = route {
                let titleChanged = editTitle != currentRoute.title
                let slugChanged = editSlug != currentRoute.slug.rawValue
                let statusChanged = editRouteStatus != currentRoute.status
                if titleChanged || slugChanged || statusChanged {
                    _ = try await CMSClient.shared.routes.update(params: UpdateRouteParams(
                        slug: currentRoute.slug,
                        title: editTitle,
                        status: editRouteStatus,
                        authorID: userID,
                        slug2: Slug(editSlug.hasPrefix("/") ? editSlug : "/" + editSlug)
                    ))
                }
            }

            // Update content status if changed
            if editContentStatus != content.status {
                _ = try await CMSClient.shared.contentData.update(params: UpdateContentDataParams(
                    contentDataID: content.contentDataID,
                    routeID: content.routeID,
                    datatypeID: content.datatypeID,
                    authorID: userID,
                    status: editContentStatus
                ))
            }

            // Update changed content fields
            for pair in fieldPairs {
                let newValue = editFieldValues[pair.field.fieldID.rawValue] ?? ""
                let oldValue = pair.value
                if newValue != oldValue {
                    if let cf = pair.contentField {
                        _ = try await CMSClient.shared.contentFields.update(params: UpdateContentFieldParams(
                            contentFieldID: cf.contentFieldID,
                            contentDataID: ContentID(contentDataID.rawValue),
                            fieldID: pair.field.fieldID,
                            fieldValue: newValue,
                            authorID: userID
                        ))
                    } else if !newValue.isEmpty {
                        _ = try await CMSClient.shared.contentFields.create(params: CreateContentFieldParams(
                            contentDataID: ContentID(contentDataID.rawValue),
                            fieldID: pair.field.fieldID,
                            fieldValue: newValue,
                            authorID: userID
                        ))
                    }
                }
            }

            isEditing = false
            isSaving = false
            await load()
        } catch {
            isSaving = false
            saveError = error.localizedDescription
        }
    }
}
