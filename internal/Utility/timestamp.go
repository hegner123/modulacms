package utility

import (
	"strconv"
	"time"
)

func TimestampI() int64 {
	return time.Now().Unix()
}

func TimestampS() string {
    return strconv.FormatInt(time.Now().Unix(), 10)
}

func TimestampReadable() string {
    return time.Now().Format(time.RFC3339)
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


