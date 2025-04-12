package utility

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"time"
)

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

func FormatJSON(b any) (string, error) {
	formatted, err := json.MarshalIndent(b, "", "    ")
	if err != nil {
		return "", err
	}
	return string(formatted), nil
}

func MakeRandomString() string {
	t := time.Now().Unix()
	tstring := strconv.FormatInt(t, 10)
	enc := base64.RawStdEncoding.EncodeToString([]byte(tstring))
	return enc
}
