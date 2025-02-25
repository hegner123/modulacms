# Modula CMS

## scripts
- make

## ToDo
- [ ] Server: GO
- [ ] Endpoints:
    - **admin/** SSR Admin
    - **content/** JSON/Protobuf endpoints to serve Clients
    - **api/vX/** versioned DataAPI to handle db and bucket interactions
- [ ] DB: SQLite / Postgres / mysql|mariadb
- [ ] Bucket: Deployment require cloud storage endpoint
- [ ] Reverse Proxy Bucket endpoints
- [ ] CMS FrontEnd: HTML, Tailwind, HTMX, WebComponents

## Proof of concept requires
- [x] Go Server
- [x] Go DB connection
- [x] DB functions
- [ ] Handle routes for Admin
- [ ] Load and confirm bucket connection
- [ ] Render html pages and templates
- [ ] Admin authentication - oAuth?
- [ ] Local authentication
- [ ] Middleware
- [ ] Dashboard
- [ ] Route editor
- [ ] Create field
- [ ] Create elements
- [x] Connect to Bucket Storage
- [ ] Media upload
- [ ] Media Optimize
- [ ] Backup
- [ ] Restore

### Dashboard
- s3 bucket connection
- db connection, Sqlite, mysql, mariadb, postgres, etc.
### Admin Types
- Configuration
- User Management
- Route Composer
- Type Composer
- Media Gallery

### Admin Templates
- Dashboard
- Users
    - User
- Route Composer: made of composed types
    - Route
        - Get all relational types
- Type Composer: made of composed fields or composed types
    - Type
        - Get all relational fields and related types
- Field Composer: direct api access
    - Field
        - Get data for field
- Media Gallery
    - Media

## Thoughts
- Query builder for admin. Dynamically build queries to show posts and post types.
- User management
- DataAPI for user front end
- Templates for admin front end 
- WebComponents hook from go html templates for field id, etc.


## Flags
- -h Display help information about the application
- -v Print the version of the application and exit
- -q Run the application with minimal or no output
- -V|V|V|V Provide more detailed output for debugging or analysis
- -p Specify a network port
- --config Specify a configuration file
- -S Enable SSL/TLS for secure connections
- --db use a connection string to connect to a db
- --reset clear tables and reset. 

## Admin Bar Links
- Github
- Support
- Documentation
- Help
- Logout

## Sidebar Links
- Dashboard
- Routes
- Data types
- Fields
- Media
- Plugins
- Users

## Style Sections
_______________________________
|________Admin Bar_____________| 
|     |                        | 
|     |                        | 
| Left|                        | 
| Side|        Editor          |
| Bar |                        |
|     |                        |
|     |                        |
|_____|________________________|

- [ ] Admin Bar 
    - Typography - sans-serif
    - Color 
        - Background Off Black - lightest
        - Text Off white/grey
- [ ] Side Bar
    - Typography - sans-serif
    - Color 
        - Background Off Black - mid
        - Text Off white/grey
- [ ] Editor
    - Typography - sans-serif
    - Color 
        - Background Off Black - darkest
        - Text Off white/grey

## End Goal 
- **CMS deliverable as an executable with templates.**
- Require load functions to register plugins and implement configs.
- Support plugins written in Javascript.
    - Plugin API:
        - Use special keys to register.
- integration support...

