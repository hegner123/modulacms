package utility

import (
	"fmt"
	"testing"
)

type TestLog struct {
	Message string
}

func TestDebug(t *testing.T) {
	tl := TestLog{Message: "Lorem ipsum et set ced sum claude."}
	DefaultLogger.Debug("Message:", tl.Message)
}

func TestInfo(t *testing.T) {
	tl := TestLog{Message: "Lorem ipsum et set ced sum claude."}
	DefaultLogger.Info("Message:", tl.Message)
}

func TestWarn(t *testing.T) {
	err := fmt.Errorf("lorem ipsum et set ced sum claude")
	tl := TestLog{Message: "lorem ipsum et set ced sum claude"}
    DefaultLogger.Warn("Message:", err, tl.Message)
}

func TestError(t *testing.T) {
	err := fmt.Errorf("Lorem ipsum et set ced sum claude.")
    DefaultLogger.Error("message", err)
}

