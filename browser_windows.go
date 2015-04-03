package main

import (
	"os/exec"
	"time"
)

func LaunchBrowser(address string) {
	time.Sleep(1 * time.Second)
	cmd := exec.Command("C:\\Program Files\\Internet Explorer\\iexplore.exe", "http://"+address)
	cmd.Run()
}
