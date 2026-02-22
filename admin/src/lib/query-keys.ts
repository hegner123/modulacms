export const queryKeys = {
  auth: {
    all: ['auth'] as const,
    me: () => [...queryKeys.auth.all, 'me'] as const,
  },
  datatypes: {
    all: ['datatypes'] as const,
    list: () => [...queryKeys.datatypes.all, 'list'] as const,
    detail: (id: string) => [...queryKeys.datatypes.all, id] as const,
  },
  fields: {
    all: ['fields'] as const,
    list: () => [...queryKeys.fields.all, 'list'] as const,
    detail: (id: string) => [...queryKeys.fields.all, id] as const,
  },
  datatypeFields: {
    all: ['datatypeFields'] as const,
    list: () => [...queryKeys.datatypeFields.all, 'list'] as const,
    byDatatype: (id: string) => [...queryKeys.datatypeFields.all, 'byDatatype', id] as const,
  },
  contentData: {
    all: ['contentData'] as const,
    list: () => [...queryKeys.contentData.all, 'list'] as const,
    detail: (id: string) => [...queryKeys.contentData.all, id] as const,
  },
  contentFields: {
    all: ['contentFields'] as const,
    list: () => [...queryKeys.contentFields.all, 'list'] as const,
    detail: (id: string) => [...queryKeys.contentFields.all, id] as const,
  },
  routes: {
    all: ['routes'] as const,
    list: () => [...queryKeys.routes.all, 'list'] as const,
    detail: (id: string) => [...queryKeys.routes.all, id] as const,
  },
  adminRoutes: {
    all: ['adminRoutes'] as const,
    list: () => [...queryKeys.adminRoutes.all, 'list'] as const,
    ordered: () => [...queryKeys.adminRoutes.all, 'ordered'] as const,
    detail: (slug: string) => [...queryKeys.adminRoutes.all, slug] as const,
  },
  adminTree: {
    all: ['adminTree'] as const,
    bySlug: (slug: string) => [...queryKeys.adminTree.all, slug] as const,
  },
  tree: {
    all: ['tree'] as const,
    bySlug: (slug: string) => [...queryKeys.tree.all, slug] as const,
  },
  adminContentData: {
    all: ['adminContentData'] as const,
    list: () => [...queryKeys.adminContentData.all, 'list'] as const,
    detail: (id: string) => [...queryKeys.adminContentData.all, id] as const,
  },
  adminContentFields: {
    all: ['adminContentFields'] as const,
    list: () => [...queryKeys.adminContentFields.all, 'list'] as const,
  },
  adminDatatypes: {
    all: ['adminDatatypes'] as const,
    list: () => [...queryKeys.adminDatatypes.all, 'list'] as const,
    detail: (id: string) => [...queryKeys.adminDatatypes.all, id] as const,
  },
  adminFields: {
    all: ['adminFields'] as const,
    list: () => [...queryKeys.adminFields.all, 'list'] as const,
    detail: (id: string) => [...queryKeys.adminFields.all, id] as const,
  },
  adminDatatypeFields: {
    all: ['adminDatatypeFields'] as const,
    list: () => [...queryKeys.adminDatatypeFields.all, 'list'] as const,
    byDatatype: (id: string) => [...queryKeys.adminDatatypeFields.all, 'byDatatype', id] as const,
  },
  media: {
    all: ['media'] as const,
    list: () => [...queryKeys.media.all, 'list'] as const,
    detail: (id: string) => [...queryKeys.media.all, id] as const,
  },
  users: {
    all: ['users'] as const,
    list: () => [...queryKeys.users.all, 'list'] as const,
    detail: (id: string) => [...queryKeys.users.all, id] as const,
  },
  roles: {
    all: ['roles'] as const,
    list: () => [...queryKeys.roles.all, 'list'] as const,
  },
  permissions: {
    all: ['permissions'] as const,
    list: () => [...queryKeys.permissions.all, 'list'] as const,
  },
  rolePermissions: {
    all: ['rolePermissions'] as const,
    list: () => [...queryKeys.rolePermissions.all, 'list'] as const,
  },
  tokens: {
    all: ['tokens'] as const,
    list: () => [...queryKeys.tokens.all, 'list'] as const,
  },
  sshKeys: {
    all: ['sshKeys'] as const,
    list: () => [...queryKeys.sshKeys.all, 'list'] as const,
  },
  audit: {
    all: ['audit'] as const,
    list: () => [...queryKeys.audit.all, 'list'] as const,
  },
  plugins: {
    all: ['plugins'] as const,
    list: () => [...queryKeys.plugins.all, 'list'] as const,
    detail: (name: string) => [...queryKeys.plugins.all, name] as const,
  },
  pluginRoutes: {
    all: ['pluginRoutes'] as const,
    list: () => [...queryKeys.pluginRoutes.all, 'list'] as const,
  },
  pluginHooks: {
    all: ['pluginHooks'] as const,
    list: () => [...queryKeys.pluginHooks.all, 'list'] as const,
  },
  fieldTypes: {
    all: ['fieldTypes'] as const,
    list: () => [...queryKeys.fieldTypes.all, 'list'] as const,
  },
  adminFieldTypes: {
    all: ['adminFieldTypes'] as const,
    list: () => [...queryKeys.adminFieldTypes.all, 'list'] as const,
  },
} as const
