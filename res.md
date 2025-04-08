# Reorganizing Go Packages for Best Practices

## Current Structure Issues

1. **Inconsistent Package Naming**
   - Non-idiomatic capitalized packages (`Auth`, `Db`, `Utility`)
   - Package paths like `github.com/hegner123/modulacms/internal/Auth` should be lowercase

2. **Interface Duplication**
   - `internal/Db/db.go` repeats methods for different database drivers
   - Over 300 methods in a single interface (`DbDriver`)

3. **Poor Separation of Concerns**
   - Authentication logic mixed with HTTP handlers
   - Database layer contains domain-specific knowledge
   - Utility package combines unrelated functions (logging, timestamps, AWS)

## Recommended Structure

```
/internal
  /app                # Application orchestration
  /auth               # Authentication services
    /oauth            # OAuth functionality
    /session          # Session management
    /token            # Token handling
  /config             # Configuration management
  /domain             # Domain models
    /content          # Content models
    /datatype         # Datatype models
    /media            # Media models
    /user             # User models
  /storage            # Storage interfaces
    /database         # Database interfaces
      /sqlite         # SQLite implementation
      /mysql          # MySQL implementation
      /postgres       # PostgreSQL implementation
    /file             # File storage
    /s3               # S3 storage
  /transport          # External communication
    /http             # HTTP handlers
  /util               # Utilities (broken into subpackages)
    /log              # Logging
    /time             # Time utilities
```

## Implementation Examples

### Domain Models (User Example):
```go
// internal/domain/user/user.go
package user

type User struct {
    ID        int64
    Email     string
    Username  string
    Hash      string
    FirstName string
    LastName  string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Repository defines user storage interface
type Repository interface {
    Find(id int64) (*User, error)
    FindByEmail(email string) (*User, error)
    Create(user *User) error
    Update(user *User) error
    Delete(id int64) error
}
```

### Database Interface:
```go
// internal/storage/database/driver.go
package database

type Driver interface {
    Connect(ctx context.Context) error
    Close() error
    Execute(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
    Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
    Begin(ctx context.Context) (Transaction, error)
}
```

This structure improves maintainability, enables proper testing through dependency injection, and follows Go conventions for package organization.
