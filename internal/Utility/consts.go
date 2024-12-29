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



type ANSIColor string

const (
	RED    ANSIColor = "\033[31m"
	GREEN  ANSIColor = "\033[32m"
	YELLOW ANSIColor = "\033[33m"
	BLUE   ANSIColor = "\033[34m"
	RESET  ANSIColor = "\033[0m" // Reset
)
