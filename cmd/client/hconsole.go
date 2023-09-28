//go:build !windows
// +build !windows

package main

var HConsole = func() int {
	return 0
}
