// +build !windows

package main

import "syscall"

func init() {
	signals = append(signals,
		syscall.SIGABRT,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
	)
}
