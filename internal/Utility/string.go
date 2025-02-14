package utility

import "strconv"

func TrimStringEnd(str string, l int) string {
	if len(str) > 0 {
		newStr := str[:len(str)-l]
		return newStr
	} else {
		return str
	}
}

func IsInt(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
