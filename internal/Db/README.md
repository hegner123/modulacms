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

1. User
2. Route
3. Field
4. Media
5. MediaDimension
6. Token
7. Session
8. Role
9. Permission
10. Datatype
11. ContentData
12. ContentField
13. AdminRoute
14. AdminField
15. AdminDatatype
16. AdminContentData
17. AdminContentField
18. Table
19. UserOauth

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