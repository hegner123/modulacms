package utility

import (
	"testing"
	"time"
)

func TestTimestampInt(t *testing.T) {
	ti := TimestampI()
	if ti <= 0 {
		t.Errorf("TimestampI returned %d, expected a positive number", ti)
	}
}

func TestTimestampString(t *testing.T) {
	ts := TimestampS()
	if ts == "" {
		t.Errorf("TimestampS returned empty string, expected a timestamp")
	}
}

func TestTimestampReadable(t *testing.T) {
	tr := TimestampReadable()
	if tr == "" {
		t.Errorf("TimestampReadable returned empty string, expected a timestamp")
	}
	
	// Verify it can be parsed as RFC3339
	_, err := time.Parse(time.RFC3339, tr)
	if err != nil {
		t.Errorf("TimestampReadable returned %s, which is not in RFC3339 format", tr)
	}
}

func TestFormatTimestampForDB(t *testing.T) {
	testTime := time.Date(2023, 4, 15, 12, 30, 45, 0, time.UTC)
	
	tests := []struct {
		name         string
		dbDriverType DbDriverType
		want         string
	}{
		{
			name:         "SQLite format",
			dbDriverType: DbSqlite,
			want:         "2023-04-15T12:30:45Z",
		},
		{
			name:         "MySQL format",
			dbDriverType: DbMysql,
			want:         "2023-04-15 12:30:45",
		},
		{
			name:         "PostgreSQL format",
			dbDriverType: DbPsql,
			want:         "2023-04-15 12:30:45",
		},
		{
			name:         "Default format for unknown driver",
			dbDriverType: "unknown",
			want:         "2023-04-15T12:30:45Z",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTimestampForDB(testTime, tt.dbDriverType)
			if got != tt.want {
				t.Errorf("FormatTimestampForDB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatTimestampForDriverString(t *testing.T) {
	testTime := time.Date(2023, 4, 15, 12, 30, 45, 0, time.UTC)
	
	tests := []struct {
		name     string
		dbDriver string
		want     string
	}{
		{
			name:     "SQLite format",
			dbDriver: "sqlite",
			want:     "2023-04-15T12:30:45Z",
		},
		{
			name:     "MySQL format",
			dbDriver: "mysql",
			want:     "2023-04-15 12:30:45",
		},
		{
			name:     "PostgreSQL format",
			dbDriver: "postgres",
			want:     "2023-04-15 12:30:45",
		},
		{
			name:     "Default format for unknown driver",
			dbDriver: "unknown",
			want:     "2023-04-15T12:30:45Z",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTimestampForDriverString(testTime, tt.dbDriver)
			if got != tt.want {
				t.Errorf("FormatTimestampForDriverString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurrentTimestampForDB(t *testing.T) {
	// This test just verifies the function runs successfully
	// and returns a non-empty string for each database type
	for _, driver := range []DbDriverType{DbSqlite, DbMysql, DbPsql} {
		result := CurrentTimestampForDB(driver)
		if result == "" {
			t.Errorf("CurrentTimestampForDB(%v) returned empty string", driver)
		}
	}
}

func TestCurrentTimestampForDriverString(t *testing.T) {
	// This test just verifies the function runs successfully
	// and returns a non-empty string for each database type
	for _, driver := range []string{"sqlite", "mysql", "postgres"} {
		result := CurrentTimestampForDriverString(driver)
		if result == "" {
			t.Errorf("CurrentTimestampForDriverString(%v) returned empty string", driver)
		}
	}
}

func TestParseDBTimestamp(t *testing.T) {
	// Test parsing timestamps from different database formats
	
	tests := []struct {
		name         string
		timestamp    string
		dbDriverType DbDriverType
		wantYear     int
		wantMonth    time.Month
		wantDay      int
		wantHour     int
		wantMin      int
		wantSec      int
		wantErr      bool
	}{
		{
			name:         "SQLite ISO8601 format",
			timestamp:    "2023-04-15T12:30:45Z",
			dbDriverType: DbSqlite,
			wantYear:     2023,
			wantMonth:    4,
			wantDay:      15,
			wantHour:     12,
			wantMin:      30,
			wantSec:      45,
			wantErr:      false,
		},
		{
			name:         "SQLite Unix timestamp",
			timestamp:    "1681563045", // 2023-04-15 12:30:45 UTC
			dbDriverType: DbSqlite,
			// Don't check exact values since they'll be converted to local timezone
			// We'll just verify the function successfully returns a time value
			wantErr:      false,
		},
		{
			name:         "MySQL format",
			timestamp:    "2023-04-15 12:30:45",
			dbDriverType: DbMysql,
			wantYear:     2023,
			wantMonth:    4,
			wantDay:      15,
			wantHour:     12,
			wantMin:      30,
			wantSec:      45,
			wantErr:      false,
		},
		{
			name:         "PostgreSQL format",
			timestamp:    "2023-04-15 12:30:45",
			dbDriverType: DbPsql,
			wantYear:     2023,
			wantMonth:    4,
			wantDay:      15,
			wantHour:     12,
			wantMin:      30,
			wantSec:      45,
			wantErr:      false,
		},
		{
			name:         "Invalid format",
			timestamp:    "not-a-timestamp",
			dbDriverType: DbSqlite,
			wantErr:      true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDBTimestamp(tt.timestamp, tt.dbDriverType)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDBTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr {
				return
			}
			
			// For Unix timestamps we don't check specific values due to timezone differences
			if tt.name != "SQLite Unix timestamp" {
				if got.Year() != tt.wantYear ||
					got.Month() != tt.wantMonth ||
					got.Day() != tt.wantDay ||
					got.Hour() != tt.wantHour ||
					got.Minute() != tt.wantMin ||
					got.Second() != tt.wantSec {
					t.Errorf("ParseDBTimestamp() = %v, want %d-%02d-%02d %02d:%02d:%02d",
						got, tt.wantYear, tt.wantMonth, tt.wantDay, tt.wantHour, tt.wantMin, tt.wantSec)
				}
			}
		})
	}
}

func TestParseDBTimestampString(t *testing.T) {
	tests := []struct {
		name         string
		timestamp    string
		dbDriver     string
		wantYear     int
		wantMonth    time.Month
		wantDay      int
		wantHour     int
		wantMin      int
		wantSec      int
		wantErr      bool
	}{
		{
			name:         "SQLite ISO8601 format",
			timestamp:    "2023-04-15T12:30:45Z",
			dbDriver:     "sqlite",
			wantYear:     2023,
			wantMonth:    4,
			wantDay:      15,
			wantHour:     12,
			wantMin:      30,
			wantSec:      45,
			wantErr:      false,
		},
		{
			name:         "MySQL format",
			timestamp:    "2023-04-15 12:30:45",
			dbDriver:     "mysql",
			wantYear:     2023,
			wantMonth:    4,
			wantDay:      15,
			wantHour:     12,
			wantMin:      30,
			wantSec:      45,
			wantErr:      false,
		},
		{
			name:         "Invalid format",
			timestamp:    "not-a-timestamp",
			dbDriver:     "sqlite",
			wantErr:      true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDBTimestampString(tt.timestamp, tt.dbDriver)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDBTimestampString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr {
				return
			}
			
			if got.Year() != tt.wantYear ||
				got.Month() != tt.wantMonth ||
				got.Day() != tt.wantDay ||
				got.Hour() != tt.wantHour ||
				got.Minute() != tt.wantMin ||
				got.Second() != tt.wantSec {
				t.Errorf("ParseDBTimestampString() = %v, want %d-%02d-%02d %02d:%02d:%02d",
					got, tt.wantYear, tt.wantMonth, tt.wantDay, tt.wantHour, tt.wantMin, tt.wantSec)
			}
		})
	}
}
