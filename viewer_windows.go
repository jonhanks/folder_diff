package main

import (
	"os"
	"os/exec"
)

func ViewImage(fname string) {
	win := os.Getenv("WINDIR")
	if win == "" {
		win = "C:\\Windows"
	}
	cmd := exec.Command(win+"\\system32\\mspaint.exe", fname)
	cmd.Run()
}
