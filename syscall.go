//go:build !windows

package tui

import "syscall"

func sysRead(fd int, p []byte) (int, error)   { return syscall.Read(fd, p) }
func setNonBlock(fd int, nonblock bool) error { return syscall.SetNonblock(fd, nonblock) }
