# Custom Admin Interfaces

ModulaCMS runs two independent content systems in parallel -- one for your site's public content and one for the admin panel itself.

## Two Content Systems

Every ModulaCMS instance manages two sets of content:

- **Public content** stores the pages, posts, and data that your site serves to visitors. This is the content your frontend application fetches and renders.
- **Admin content** stores the structure and configuration of the admin panel. The built-in admin panel at `/admin/` reads from admin content to determine which screens, forms, and navigation items to display.

These are not stages in a publishing pipeline. Public content and admin content are fully independent systems that happen to share the same schema design -- datatypes, fields, content entries, and routes. Each has its own publish flow, versioning, and tree structure.

## The built-in admin panel

ModulaCMS ships with a server-rendered admin panel that covers content management, schema editing, media uploads, user administration, and settings. This panel is powered by admin content -- it reads admin datatypes, admin fields, and admin content entries to build its interface.

You don't need to replace the built-in panel. For most projects, it handles everything out of the box.

## Build your own admin panel

When you need a custom admin experience -- a React dashboard, a mobile management app, a specialized editorial workflow -- you build it against the admin content API. The admin API exposes the same CRUD operations as the public content API, prefixed with `admin`:

| Public endpoint | Admin equivalent |
|-----------------|-----------------|
| `/api/v1/contentdata` | `/api/v1/admincontentdatas` |
| `/api/v1/contentfields` | `/api/v1/admincontentfields` |
| `/api/v1/datatype` | `/api/v1/admindatatypes` |
| `/api/v1/fields` | `/api/v1/adminfields` |
| `/api/v1/fieldtypes` | `/api/v1/adminfieldtypes` |
| `/api/v1/routes` | `/api/v1/adminroutes` |

Create admin datatypes to define new admin screens. Add admin fields to control what data those screens collect. Populate admin content to fill them. Your custom admin panel reads and writes through these endpoints, and you can change its structure at runtime without code deploys.

> **Good to know**: Any feature that works on public content works identically on admin content. Tree structures, field types, batch updates, and version history all apply to both systems.

## What this means in practice

Your admin panel is itself managed through the CMS. Want to add a new screen to your admin interface? Create an admin datatype, attach fields, and create admin content entries -- all through API calls. Want to reorganize your admin navigation? Update the admin content tree. No redeployment, no code changes.

This separation also means you can version, audit, and roll back admin panel changes the same way you manage site content.

## Next steps

- [Build a custom admin interface](building-interfaces.md) -- step-by-step guide to creating admin screens via the API
- [Authentication and access control](authentication.md) -- log in, manage users, and control permissions
