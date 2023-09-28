//go:build windows
// +build windows

package main

import (
	"syscall"
	"unsafe"
)

var HConsole = func() int {

	FreeConsole := syscall.NewLazyDLL("kernel32.dll").NewProc("FreeConsole")
	FreeConsole.Call()

	FindWindowA := syscall.NewLazyDLL("user32.dll").NewProc("FindWindowA")
	lpClassName, _ := syscall.BytePtrFromString("ConsoleWindowClass")
	Stealth, _, _ := FindWindowA.Call(uintptr(unsafe.Pointer(lpClassName)), 0)
	ShowWindow := syscall.NewLazyDLL("user32.dll").NewProc("ShowWindow")
	ShowWindow.Call(Stealth, 0)

	getWin := syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleWindow")
	//ShowWindow
	showWin := syscall.NewLazyDLL("user32.dll").NewProc("ShowWindow")
	hwnd, _, _ := getWin.Call()
	if hwnd == 0 {
		return 1
	}
	showWin.Call(hwnd, 0)

	getWin = syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleWindow")
	hwnd, _, _ = getWin.Call()
	if hwnd != 0 {
		showWindowAsync := syscall.NewLazyDLL("user32.dll").NewProc("ShowWindowAsync")
		showWindowAsync.Call(hwnd, 0)
	}
	return 0
}
