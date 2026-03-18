# Custom Admin Interfaces

ModulaCMS runs two independent content systems in parallel -- one for your site's public content and one for managing your admin panel interface.

## Two content systems

Every ModulaCMS instance manages two sets of content:

- **Public content** stores the pages, posts, and data that your site serves to visitors. This is the content your frontend application fetches and renders.
- **Admin content** stores the structure, configuration, and data that powers your admin panel interface. It uses the same schema design as public content -- datatypes, fields, content entries, and routes -- but is completely independent.

These are not stages in a publishing pipeline. Public content and admin content are fully independent systems that happen to share the same schema design. Each has its own publish flow, versioning, and tree structure.

The key idea: customizing and extending your admin panel works exactly the same way as maintaining your client's content. Same tools, same API, same workflow.

## The built-in admin panel

ModulaCMS ships with a server-rendered admin panel at `/admin/` that covers content management, schema editing, media uploads, user administration, and settings. This is a static, developer-focused panel built with HTMX and templ -- a bare-minimum reference implementation that directly manages both content systems. It provides stock feature control with no customization.

The built-in panel is your starting point. It gives you full control over public content and admin content tables out of the box.

## Pre-built admin panels

ModulaCMS will offer fully featured admin panels paired with content models designed for the admin content tables. Start with a pre-built layout, then customize and extend it as easily as you maintain your client's content -- add screens, rearrange navigation, modify forms, all at runtime through the admin content API.

## The admin content API

The admin API exposes the same CRUD operations as the public content API, prefixed with `admin`:

| Public endpoint | Admin equivalent |
|-----------------|-----------------|
| `/api/v1/contentdata` | `/api/v1/admincontentdatas` |
| `/api/v1/contentfields` | `/api/v1/admincontentfields` |
| `/api/v1/datatype` | `/api/v1/admindatatypes` |
| `/api/v1/fields` | `/api/v1/adminfields` |
| `/api/v1/fieldtypes` | `/api/v1/adminfieldtypes` |
| `/api/v1/routes` | `/api/v1/adminroutes` |

Create admin datatypes to define screens in your admin panel. Add admin fields to control what data those screens collect. Populate admin content to fill them. Whether you start from a pre-built panel or build from scratch, your admin panel reads and writes through these endpoints, and you can change its structure at runtime without code deploys.

> **Good to know**: Any feature that works on public content works identically on admin content. Tree structures, field types, batch updates, and version history all apply to both systems.

## What this means in practice

Want to add a new screen to your admin panel? Create an admin datatype, attach fields, and create admin content entries -- all through API calls. Want to reorganize navigation? Update the admin content tree. No redeployment, no code changes.

This separation also means you can version, audit, and roll back admin panel changes the same way you manage site content.

## Next steps

- [Build a custom admin interface](building-interfaces.md) -- step-by-step guide to creating admin screens via the API
- [Authentication and access control](authentication.md) -- log in, manage users, and control permissions
