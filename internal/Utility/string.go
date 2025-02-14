package utility

func TrimStringEnd(str string, l int) string {
	if len(str) > 0 {
		newStr := str[:len(str)-l]
		return newStr
	} else {
		return str
	}
}
