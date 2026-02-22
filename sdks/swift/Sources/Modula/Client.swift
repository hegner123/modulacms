import Foundation

public struct ClientConfig: Sendable {
    public let baseURL: String
    public let apiKey: String
    public let urlSession: URLSession?

    public init(baseURL: String, apiKey: String = "", urlSession: URLSession? = nil) {
        self.baseURL = baseURL
        self.apiKey = apiKey
        self.urlSession = urlSession
    }
}

public final class ModulaClient: Sendable {
    // Standard CRUD resources
    public let contentData: Resource<ContentData, CreateContentDataParams, UpdateContentDataParams, ContentID>
    public let contentFields: Resource<ContentField, CreateContentFieldParams, UpdateContentFieldParams, ContentFieldID>
    public let contentRelations: Resource<ContentRelation, CreateContentRelationParams, UpdateContentRelationParams, ContentRelationID>
    public let datatypes: Resource<Datatype, CreateDatatypeParams, UpdateDatatypeParams, DatatypeID>
    public let datatypeFields: Resource<DatatypeField, CreateDatatypeFieldParams, UpdateDatatypeFieldParams, DatatypeFieldID>
    public let fields: Resource<Field, CreateFieldParams, UpdateFieldParams, FieldID>
    public let media: Resource<Media, NoCreate, UpdateMediaParams, MediaID>
    public let mediaDimensions: Resource<MediaDimension, CreateMediaDimensionParams, UpdateMediaDimensionParams, MediaDimensionID>
    public let routes: Resource<Route, CreateRouteParams, UpdateRouteParams, RouteID>
    public let roles: Resource<Role, CreateRoleParams, UpdateRoleParams, RoleID>
    public let permissions: Resource<Permission, CreatePermissionParams, UpdatePermissionParams, PermissionID>
    public let users: Resource<User, CreateUserParams, UpdateUserParams, UserID>
    public let tokens: Resource<Token, CreateTokenParams, UpdateTokenParams, TokenID>
    public let usersOauth: Resource<UserOauth, CreateUserOauthParams, UpdateUserOauthParams, UserOauthID>
    public let tables: Resource<Table, CreateTableParams, UpdateTableParams, TableID>

    // Admin CRUD resources
    public let adminContentData: Resource<AdminContentData, CreateAdminContentDataParams, UpdateAdminContentDataParams, AdminContentID>
    public let adminContentFields: Resource<AdminContentField, CreateAdminContentFieldParams, UpdateAdminContentFieldParams, AdminContentFieldID>
    public let adminDatatypes: Resource<AdminDatatype, CreateAdminDatatypeParams, UpdateAdminDatatypeParams, AdminDatatypeID>
    public let adminDatatypeFields: Resource<AdminDatatypeField, CreateAdminDatatypeFieldParams, UpdateAdminDatatypeFieldParams, AdminDatatypeFieldID>
    public let adminFields: Resource<AdminField, CreateAdminFieldParams, UpdateAdminFieldParams, AdminFieldID>
    public let adminRoutes: Resource<AdminRoute, CreateAdminRouteParams, UpdateAdminRouteParams, AdminRouteID>
    public let fieldTypes: Resource<FieldTypeInfo, CreateFieldTypeParams, UpdateFieldTypeParams, FieldTypeID>
    public let adminFieldTypes: Resource<AdminFieldTypeInfo, CreateAdminFieldTypeParams, UpdateAdminFieldTypeParams, AdminFieldTypeID>

    // Specialized resources
    public let auth: AuthResource
    public let mediaUpload: MediaUploadResource
    public let adminTree: AdminTreeResource
    public let content: ContentDeliveryResource
    public let sshKeys: SSHKeysResource
    public let sessions: SessionsResource
    public let importResource: ImportResource
    public let contentBatch: ContentBatchResource

    // RBAC resources
    public let rolePermissions: RolePermissionsResource

    // Plugin resources
    public let plugins: PluginsResource
    public let pluginRoutes: PluginRoutesResource
    public let pluginHooks: PluginHooksResource

    // Config resource
    public let config: ConfigResource

    public init(config: ClientConfig) throws {
        guard !config.baseURL.isEmpty else {
            throw APIError(statusCode: 0, message: "modula: BaseURL is required")
        }

        var baseURL = config.baseURL
        while baseURL.hasSuffix("/") {
            baseURL.removeLast()
        }

        let session: URLSession
        if let provided = config.urlSession {
            session = provided
        } else {
            let configuration = URLSessionConfiguration.default
            configuration.timeoutIntervalForRequest = 30
            configuration.timeoutIntervalForResource = 300
            configuration.httpCookieAcceptPolicy = .never
            configuration.httpShouldSetCookies = false
            configuration.httpCookieStorage = nil
            session = URLSession(configuration: configuration)
        }

        let http = HTTPClient(baseURL: baseURL, apiKey: config.apiKey, session: session)

        // Standard CRUD
        contentData = Resource(path: "/api/v1/contentdata", http: http)
        contentFields = Resource(path: "/api/v1/contentfields", http: http)
        contentRelations = Resource(path: "/api/v1/contentrelations", http: http)
        datatypes = Resource(path: "/api/v1/datatype", http: http)
        datatypeFields = Resource(path: "/api/v1/datatypefields", http: http)
        fields = Resource(path: "/api/v1/fields", http: http)
        media = Resource(path: "/api/v1/media", http: http)
        mediaDimensions = Resource(path: "/api/v1/mediadimensions", http: http)
        routes = Resource(path: "/api/v1/routes", http: http)
        roles = Resource(path: "/api/v1/roles", http: http)
        permissions = Resource(path: "/api/v1/permissions", http: http)
        users = Resource(path: "/api/v1/users", http: http)
        tokens = Resource(path: "/api/v1/tokens", http: http)
        usersOauth = Resource(path: "/api/v1/usersoauth", http: http)
        tables = Resource(path: "/api/v1/tables", http: http)

        // Admin CRUD
        adminContentData = Resource(path: "/api/v1/admincontentdatas", http: http)
        adminContentFields = Resource(path: "/api/v1/admincontentfields", http: http)
        adminDatatypes = Resource(path: "/api/v1/admindatatypes", http: http)
        adminDatatypeFields = Resource(path: "/api/v1/admindatatypefields", http: http)
        adminFields = Resource(path: "/api/v1/adminfields", http: http)
        adminRoutes = Resource(path: "/api/v1/adminroutes", http: http)
        fieldTypes = Resource(path: "/api/v1/fieldtypes", http: http)
        adminFieldTypes = Resource(path: "/api/v1/adminfieldtypes", http: http)

        // Specialized
        auth = AuthResource(http: http)
        mediaUpload = MediaUploadResource(http: http)
        adminTree = AdminTreeResource(http: http)
        content = ContentDeliveryResource(http: http)
        sshKeys = SSHKeysResource(http: http)
        sessions = SessionsResource(http: http)
        importResource = ImportResource(http: http)
        contentBatch = ContentBatchResource(http: http)

        // RBAC
        rolePermissions = RolePermissionsResource(http: http)

        // Plugin
        plugins = PluginsResource(http: http)
        pluginRoutes = PluginRoutesResource(http: http)
        pluginHooks = PluginHooksResource(http: http)

        // Config
        self.config = ConfigResource(http: http)
    }
}
