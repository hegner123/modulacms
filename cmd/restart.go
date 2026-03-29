package main

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/hegner123/modulacms/internal/utility"
)

// restartCh is a package-level channel that triggers a graceful restart.
// The admin panel handler sends on this channel; serve.go reads from it.
var restartCh = make(chan struct{}, 1)

// RequestRestart signals the serve loop to perform a graceful restart.
// Non-blocking: if a restart is already pending the call is a no-op.
func RequestRestart() {
	select {
	case restartCh <- struct{}{}:
	default:
	}
}

// execSelf replaces the current process with a new instance of the same
// binary using the same arguments and environment. This is the standard
// Unix pattern for in-place restart. On success this function never returns.
func execSelf() error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return err
	}

	utility.DefaultLogger.Info("Restarting process", "binary", execPath)
	return syscall.Exec(execPath, os.Args, os.Environ())
}
