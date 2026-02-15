package utility

// StorageUnit represents a digital storage size unit (KB, MB, GB, TB).
type StorageUnit int64

// Standard storage unit constants.
const (
	KB StorageUnit = 1 << 10
	MB StorageUnit = 1 << 20
	GB StorageUnit = 1 << 30
	TB StorageUnit = 1 << 40
)

// Standard MIME types.
const (
    AppJson string = "application/json"
)

// SizeInBytes converts a value in the given StorageUnit to bytes.
func SizeInBytes(value int64, unit StorageUnit) int64 {
	return value * int64(unit)
}

// LogLevelStyle defines the style for a specific log level
type LogLevelStyle struct {
	LevelName string
	Style     func(string) string
}
