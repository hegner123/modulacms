package utility

type StorageUnit int64

const (
	KB StorageUnit = 1 << 10
	MB StorageUnit = 1 << 20
	GB StorageUnit = 1 << 30
	TB StorageUnit = 1 << 40
)

const (
    AppJson string = "application/json"
)


func SizeInBytes(value int64, unit StorageUnit) int64 {
	return value * int64(unit)
}

// LogLevelStyle defines the style for a specific log level
type LogLevelStyle struct {
	LevelName string
	Style     func(string) string
}
