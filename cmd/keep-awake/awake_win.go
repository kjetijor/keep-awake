//go:build windows
// +build windows

package main

import (
	"log"
	"strings"
	"syscall"
)

const (
	EsSystemRequired = 0x00000001
	EsContinuous     = 0x80000000
)

var kernel32 *syscall.LazyDLL
var setThreadExecStateProc *syscall.LazyProc

func init() {
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	if kernel32 == nil {
		log.Panicf("failed to load kernel32.dll")
	}
	setThreadExecStateProc = kernel32.NewProc("SetThreadExecutionState")
	if setThreadExecStateProc == nil {
		log.Panicf("failed to load SetThreadExecutionState")
	}
}

// The operation completed successfully
func completed_successfully(err error) bool {
	if strings.Contains(err.Error(), "The operation completed successfully") {
		return true
	}
	return false
}

func keep_awake() {
	if _, _, err := setThreadExecStateProc.Call(uintptr(EsSystemRequired)); err != nil && !completed_successfully(err) {

		log.Fatalf("failed to call setThreadExecStateProc with %v", err)
	}
}
