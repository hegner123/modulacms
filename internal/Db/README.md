# Database By Resource

This directory contains database functions organized by resource, making it easier to maintain and update database operations for specific resources.

## Directory Structure

- `db.go`: Main database driver interfaces and connection methods
- `main.go`: Utility functions used across resource files
- `imports.go`: String struct definitions for all resources
- Resource files: Individual Go files for each resource, containing CRUD operations

## Resources

Each resource file follows this structure:

1. **Structs Section**: Contains struct definitions for:
   - Main resource struct
   - CreateParams struct
   - UpdateParams struct
   - HistoryEntry struct
   - Form parameter structs

2. **Generic Section**: Contains generic mapping functions:
   - MapCreateParams
   - MapUpdateParams
   - MapString functions

3. **SQLite Section**: Contains SQLite-specific implementations:
   - Map functions
   - CRUD operations

4. **MySQL Section**: Contains MySQL-specific implementations:
   - Map functions
   - CRUD operations

5. **PostgreSQL Section**: Contains PostgreSQL-specific implementations:
   - Map functions
   - CRUD operations

## Resources Included

The following resources are included in this organization:

1. Permission
2. Role
3. User
4. UserOauth
5. Table
6. Route
7. AdminRoute
8. Datatype
9. Field
10. AdminDatatype
11. AdminField
12. Media
13. MediaDimension
14. Token
15. Session
16. ContentData
17. ContentField
18. AdminContentData
19. AdminContentField
20. DatatypesFields
20. AdminDatatypesFields

## Usage

To use the database functions, import the package and use the appropriate database driver:

```go
import (
    "github.com/hegner123/modulacms/internal/Db/db_by_resource"
    config "github.com/hegner123/modulacms/internal/Config"
)

func main() {
    cfg := config.Config{
        DbPath: "path/to/database.db",
    }
    
    // Create a new database connection
    db, err := db.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    
    // Initialize the database
    err = db.InitDb(nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Use the database functions
    users, err := db.ListUsers()
    if err != nil {
        log.Fatal(err)
    }
    
    // Use the database functions for a specific resource
    user, err := db.GetUser(1)
    if err != nil {
        log.Fatal(err)
    }
}
```

## Development

When adding a new resource or updating an existing one, follow these steps:

1. Create or update the struct definitions in the appropriate resource file
2. Implement the required mapping functions
3. Implement the CRUD operations for each database type (SQLite, MySQL, PostgreSQL)
4. Update the imports.go file with any necessary string struct definitions
5. Update the db.go file to include the new resource in the interface and CreateAllTables method
