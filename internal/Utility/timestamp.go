package utility

import (
	"strconv"
	"time"
)

// DbDriverType represents database driver types to avoid import cycle
type DbDriverType string

const (
	// Database driver constants
	DbSqlite DbDriverType = "sqlite"
	DbMysql  DbDriverType = "mysql"
	DbPsql   DbDriverType = "postgres"
)

func TimestampI() int64 {
	return time.Now().Unix()
}

func TimestampS() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

// TimestampReadable returns a time string in RFC3339 format
// This is a generic format that works across systems
func TimestampReadable() string {
	return time.Now().Format(time.RFC3339)
}

// FormatTimestampForDB returns a timestamp string in the format expected by the specified database
// Each database has a different format requirement:
// - SQLite: TEXT as ISO8601 string or Unix timestamp
// - MySQL: TIMESTAMP as 'YYYY-MM-DD HH:MM:SS'
// - PostgreSQL: TIMESTAMP as 'YYYY-MM-DD HH:MM:SS'
func FormatTimestampForDB(t time.Time, dbDriverType DbDriverType) string {
	switch dbDriverType {
	case DbMysql, DbPsql:
		return t.Format("2006-01-02 15:04:05")
	default:
		return t.Format(time.RFC3339)
	}
}

// FormatTimestampForDriverString handles a driver name as string
func FormatTimestampForDriverString(t time.Time, dbDriver string) string {
	switch dbDriver {
	case "sqlite":
		return FormatTimestampForDB(t, DbSqlite)
	case "mysql":
		return FormatTimestampForDB(t, DbMysql)
	case "postgres":
		return FormatTimestampForDB(t, DbPsql)
	default:
		return t.Format(time.RFC3339)
	}
}

// CurrentTimestampForDB returns the current time formatted for the specified database
func CurrentTimestampForDB(dbDriverType DbDriverType) string {
	return FormatTimestampForDB(time.Now(), dbDriverType)
}

// CurrentTimestampForDriverString returns the current time formatted for the specified database driver string
func CurrentTimestampForDriverString(dbDriver string) string {
	return FormatTimestampForDriverString(time.Now(), dbDriver)
}

// ParseDBTimestamp parses a timestamp string from the database based on the database driver type
func ParseDBTimestamp(timestamp string, dbDriverType DbDriverType) (*time.Time, error) {
	var t time.Time
	var err error
	
	switch dbDriverType {
	case DbSqlite:
		// Try RFC3339 first, then Unix timestamp
		t, err = time.Parse(time.RFC3339, timestamp)
		if err != nil {
			// Try Unix timestamp
			i, err := strconv.ParseInt(timestamp, 10, 64)
			if err != nil {
				return nil, err
			}
			t = time.Unix(i, 0)
			return &t, nil
		}
	case DbMysql, DbPsql:
		t, err = time.Parse("2006-01-02 15:04:05", timestamp)
		if err != nil {
			return nil, err
		}
	default:
		t, err = time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return nil, err
		}
	}
	
	return &t, nil
}

// ParseDBTimestampString parses a timestamp with a driver name as string
func ParseDBTimestampString(timestamp string, dbDriver string) (*time.Time, error) {
	switch dbDriver {
	case "sqlite":
		return ParseDBTimestamp(timestamp, DbSqlite)
	case "mysql":
		return ParseDBTimestamp(timestamp, DbMysql)
	case "postgres":
		return ParseDBTimestamp(timestamp, DbPsql)
	default:
		return ParseDBTimestamp(timestamp, DbSqlite)
	}
}

func ParseTimeReadable(s string) *time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}

func TokenExpiredTime() (string, int64) {
	now := time.Now()
	t := now.Add(168 * time.Hour).Unix()
	return strconv.FormatInt(t, 10), t
}

func TimestampLessThan(a string) bool {
	aInt, err := strconv.ParseInt(a, 10, 64)
	if err != nil {
		return false
	}
	return aInt < TimestampI()
}
