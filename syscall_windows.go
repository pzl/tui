//go:build windows

package tui

import (
	"syscall"
)

func sysRead(fd int, p []byte) (int, error) { return syscall.Read(syscall.Handle(uintptr(fd)), p) }
func setNonBlock(fd int, nonblock bool) error {
	return syscall.SetNonblock(syscall.Handle(uintptr(fd)), nonblock)
}
