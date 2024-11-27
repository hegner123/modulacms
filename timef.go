package main

import (
	"fmt"
	"time"
)

func timestampI() int64 {
	return time.Now().Unix()
}

func timestampS() string {
	return fmt.Sprint(time.Now().Unix())
}
