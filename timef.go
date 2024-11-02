package main

import (
	"fmt"
	"time"
)

func timestamp() string {
	return fmt.Sprint(time.Now().Unix())
}
