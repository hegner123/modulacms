package utility

import (
	"fmt"
	"strconv"
	"time"
)

func TimestampI() int64 {
	return time.Now().Unix()
}

func TimestampS() string {
	return fmt.Sprint(time.Now().Unix())
}

func TokenExpiredTime() (string, int64) {
	t := time.Now().Add(168 * time.Hour).Unix()
	s := fmt.Sprint(t)
	return s, t
}

func TimestampLessThan(a string) bool {
	aInt, err := strconv.ParseInt(a, 10, 64)
	if err != nil {
		return false
	}
	now := TimestampI()

	if aInt < now {
		return true
	} else {
		return false
	}

}
