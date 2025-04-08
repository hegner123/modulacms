package utility

type StorageUnit int64

const (
	KB StorageUnit = 1 << 10
	MB StorageUnit = 1 << 20
	GB StorageUnit = 1 << 30
	TB StorageUnit = 1 << 40
)

func SizeInBytes(value int64, unit StorageUnit) int64 {
	return value * int64(unit)
}

// LogLevelStyle defines the style for a specific log level
type LogLevelStyle struct {
	LevelName string
	Style     func(string) string
}

/*

BRIGHTBlackB= "\033[100m"
BRIGHTRedB= "\033[101m"
BRIGHTGreenB= "\033[102m"
BRIGHTYellowB= "\033[103m"
BRIGHTBlueB= "\033[104m"
BRIGHTMagentaB= "\033[105m"
BRIGHTCyanB= "\033[106m"
BRIGHTWhiteB= "\033[107m"
Resetting Formatting
To reset the text formatting and colors back to the terminal’s defaults, use:

Reset= "\033[0m"
*/
