package utility

import (
	"fmt"
	"testing"
	"time"
)

func TestTimestampInt(t *testing.T) {
	fmt.Println(time.Now().String())
	ti := TimestampI()
	fmt.Println(ti)
}

func TestTimestampString(t *testing.T) {
	fmt.Println(time.Now().String())
	ti := TimestampS()
	fmt.Println(ti)
}

func TestTimestampReadable(t *testing.T) {
	fmt.Println(time.Now().String())
	ti := TimestampReadable()
	fmt.Println(ti)
}
