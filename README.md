# Modula CMS

## ToDo
- Server: GO
- API:
    - CMS API
    - FrontEnd API
- DB: SQLite / Postgres
- Bucket: Deployment require cloud storage endpoint
- CMS FrontEnd: HTML, Tailwind, HTMX, WebComponents
- DevOps: Github actions to deploy

## Proof of concept requires
- Go Server
- Go DB connection
- Handle routes for admin
- handle routes for frontend
- admin auth - oAuth?
- dashboard
- Route editor
- Create field
- media upload

Visible on admin frontend
- s3 bucket connection
- db connection, Sqlite, mysql, mariadb, postgres, etc.


## Thoughts
- Query builder for admin. Dynamically build queries to show posts and post types.
- User management
- DataAPI for user front end
- Templates for admin front end 
- WebComponents hook from go html templates for field id, etc.


## Flags
-h Display help information about the application
-v Print the version of the application and exit
-q Run the application with minimal or no output
-V|V|V|V Provide more detailed output for debugging or analysis
-p Specify a network port
--config Specify a configuration file
-S Enable SSL/TLS for secure connections
--db use a connection string to connect to a db
--reset clear tables and reset. 


## End Goal 
- **CMS deliverable as an executable with templates.**
- Require load functions to register plugins and implement configs.
- Support plugins written in Javascript.
    - Plugin API:
        - Use special keys to register.
- integration support...

