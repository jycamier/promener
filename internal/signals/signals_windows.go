//go:build windows

package signals

import (
	"os"
)

// Shutdown returns the OS signals to listen for graceful shutdown.
// On Windows, only os.Interrupt is supported (Ctrl+C, Ctrl+Break).
// Note: CTRL_CLOSE_EVENT, CTRL_LOGOFF_EVENT and CTRL_SHUTDOWN_EVENT
// are automatically mapped to syscall.SIGTERM by the Go runtime.
func Shutdown() []os.Signal {
	return []os.Signal{os.Interrupt}
}
