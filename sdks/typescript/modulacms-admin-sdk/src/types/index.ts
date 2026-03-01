/**
 * Barrel re-export of all public types from the ModulaCMS Admin SDK type system.
 *
 * @module types
 */

export type {
  Brand,
  UserID,
  AdminContentID,
  AdminContentFieldID,
  AdminContentRelationID,
  AdminContentVersionID,
  AdminDatatypeID,
  AdminFieldID,
  AdminRouteID,
  ContentID,
  ContentFieldID,
  ContentRelationID,
  ContentVersionID,
  LocaleID,
  DatatypeID,
  FieldID,
  MediaID,
  RoleID,
  FieldTypeID,
  AdminFieldTypeID,
  RouteID,
  SessionID,
  UserOauthID,
  Slug,
  Email,
  URL,
  ContentStatus,
  ContentFormat,
  FieldType,
  RequestOptions,
  ApiError,
  PaginationParams,
  PaginatedResponse,
} from './common.js'

export { isApiError } from './common.js'

export type {
  LoginRequest,
  LoginResponse,
  MeResponse,
} from './auth.js'

export type {
  AdminRoute,
  AdminContentData,
  AdminContentField,
  AdminDatatype,
  AdminField,
  AdminContentRelation,
  CreateAdminRouteParams,
  CreateAdminContentDataParams,
  CreateAdminContentFieldParams,
  CreateAdminDatatypeParams,
  CreateAdminFieldParams,
  UpdateAdminRouteParams,
  UpdateAdminContentDataParams,
  UpdateAdminContentFieldParams,
  UpdateAdminDatatypeParams,
  UpdateAdminFieldParams,
  MoveAdminContentDataParams,
  MoveAdminContentDataResponse,
  ReorderAdminContentDataParams,
  ReorderAdminContentDataResponse,
} from './admin.js'

export type {
  ContentData,
  ContentField,
  ContentRelation,
  ContentVersion,
  AdminContentVersion,
  CreateContentDataParams,
  CreateContentFieldParams,
  UpdateContentDataParams,
  UpdateContentFieldParams,
  BatchContentUpdateParams,
  BatchContentUpdateResponse,
} from './content.js'

export type {
  Datatype,
  Field,
  FieldTypeInfo,
  AdminFieldTypeInfo,
  AuthorView,
  DatatypeFullView,
  CreateDatatypeParams,
  CreateFieldParams,
  CreateFieldTypeParams,
  CreateAdminFieldTypeParams,
  UpdateDatatypeParams,
  UpdateFieldParams,
  UpdateFieldTypeParams,
  UpdateAdminFieldTypeParams,
} from './schema.js'

export type {
  Media,
  MediaDimension,
  CreateMediaParams,
  CreateMediaDimensionParams,
  UpdateMediaParams,
  UpdateMediaDimensionParams,
  MediaHealthResponse,
  MediaCleanupResponse,
} from './media.js'

export type {
  User,
  Role,
  Token,
  UserOauth,
  Session,
  SshKey,
  SshKeyListItem,
  UserWithRoleLabel,
  UserOauthView,
  UserSshKeyView,
  SessionView,
  TokenView,
  UserFullView,
  CreateUserParams,
  CreateRoleParams,
  CreateTokenParams,
  CreateUserOauthParams,
  CreateSshKeyRequest,
  UpdateUserParams,
  UpdateRoleParams,
  UpdateTokenParams,
  UpdateUserOauthParams,
  UpdateSessionParams,
} from './users.js'

export type {
  Route,
  CreateRouteParams,
  UpdateRouteParams,
} from './routing.js'

export type {
  Table,
  CreateTableParams,
  UpdateTableParams,
} from './tables.js'

export type {
  ImportFormat,
  ImportResponse,
} from './import.js'

export type {
  ContentTreeField,
  ContentTreeNode,
  AdminTreeResponse,
  TreeFormat,
} from './tree.js'
