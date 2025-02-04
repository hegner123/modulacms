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

type ANSIForegroundColor string
type ANSIBackgroundColor string
type ANSIBrightForegroundColor string
type ANSIBrightBackgroundColor string
type ANSIReset string

const (
	BLACKF         ANSIForegroundColor       = "\033[30m"
	REDF           ANSIForegroundColor       = "\033[31m"
	GREENF         ANSIForegroundColor       = "\033[32m"
	YELLOWF        ANSIForegroundColor       = "\033[33m"
	BLUEF          ANSIForegroundColor       = "\033[34m"
	MAGENTAF       ANSIForegroundColor       = "\033[35m"
	CYANF          ANSIForegroundColor       = "\033[36m"
	WHITEF         ANSIForegroundColor       = "\033[37m"
	BLACKB         ANSIBackgroundColor       = "\033[40m"
	BRIGHTBLACKF   ANSIBrightForegroundColor = "\033[90m"
	BRIGHTREDF     ANSIBrightForegroundColor = "\033[91m"
	BRIGHTGREENF   ANSIBrightForegroundColor = "\033[92m"
	BRIGHTYELLOWF  ANSIBrightForegroundColor = "\033[93m"
	BRIGHTBLUEF    ANSIBrightForegroundColor = "\033[94m"
	BRIGHTMAGENTAF ANSIBrightForegroundColor = "\033[95m"
	BRIGHTCYANF    ANSIBrightForegroundColor = "\033[96m"
	BRIGHTWHITEF   ANSIBrightForegroundColor = "\033[97m"
	REDB           ANSIBackgroundColor       = "\033[41m"
	GREENB         ANSIBackgroundColor       = "\033[42m"
	YELLOWB        ANSIBackgroundColor       = "\033[43m"
	BLUEB          ANSIBackgroundColor       = "\033[44m"
	MAGENTAB       ANSIBackgroundColor       = "\033[45m"
	CYANB          ANSIBackgroundColor       = "\033[46m"
	WHITEB         ANSIBackgroundColor       = "\033[47m"
	BRIGHTBLACKB   ANSIBrightBackgroundColor = "\033[100m"
	BRIGHTREDB     ANSIBrightBackgroundColor = "\033[101m"
	BRIGHTGREENB   ANSIBrightBackgroundColor = "\033[102m"
	BRIGHTYELLOWB  ANSIBrightBackgroundColor = "\033[103m"
	BRIGHTBLUEB    ANSIBrightBackgroundColor = "\033[104m"
	BRIGHTMAGENTAB ANSIBrightBackgroundColor = "\033[105m"
	BRIGHTCYANB    ANSIBrightBackgroundColor = "\033[106m"
	BRIGHTWHITEB   ANSIBrightBackgroundColor = "\033[107m"
	RESET          ANSIReset                 = "\033[0m" // Reset
)

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
