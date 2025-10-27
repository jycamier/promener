//go:build unix || darwin

package signals

import (
	"os"
	"syscall"
)

// Shutdown returns the OS signals to listen for graceful shutdown.
// On Unix systems, this includes SIGINT (Ctrl+C) and SIGTERM.
func Shutdown() []os.Signal {
	return []os.Signal{os.Interrupt, syscall.SIGTERM}
}
