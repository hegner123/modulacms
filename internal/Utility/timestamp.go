package utility

import (
	"fmt"
	"time"
)

func TimestampI() int64 {
	return time.Now().Unix()
}

func TimestampS() string {
	return fmt.Sprint(time.Now().Unix())
}
