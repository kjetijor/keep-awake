//go:build !windows
// +build !windows

package main

import "log"

func keep_awake() {
	log.Printf("keep awake on windows")
}
