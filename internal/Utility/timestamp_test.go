package utility

import "testing"

func TestTimestampLessThan(t *testing.T) {
	aStamp := "1738340471"
	res := TimestampLessThan(aStamp)
	if !res {
		t.FailNow()
	}

}
