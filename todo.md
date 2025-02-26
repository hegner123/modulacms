# Modula CMS

## scripts
- make

## ToDo
- [ ] Server: GO
- [ ] Endpoints:
    - **api/vX/admin/** JSON endpoints to serve admin
    - **api/vX/client/** JSON endpoints to serve Clients
- [ ] DB: SQLite 
- [ ] Bucket: Deployment require cloud storage endpoint
- [ ] Reverse Proxy Bucket endpoints

## Proof of concept requires
- [x] Go Server
- [x] Go DB connection
- [x] DB CRUD functions
- [ ] Handle routes for Admin api
- [ ] Handle routes for Client api
- [x] Load and confirm bucket connection
- [ ] Admin authentication - oAuth?
- [ ] Local authentication
- [ ] Middleware
- [x] Connect to Bucket Storage
- [x] Media upload
- [ ] Media Optimize
- [x] Backup
- [ ] Restore


## Flags
- -h Display help information about the application
- -v Print the version of the application and exit
- -q Run the application with minimal or no output
- -V|V|V|V Provide more detailed output for debugging or analysis
- -p Specify a network port
- -config Specify a configuration file
- -S Enable SSL/TLS for secure connections
- -db use a connection string to connect to a db
- -reset clear tables and reset. 




## End Goal 
[ ] **CMS deliverable as an executable.**

