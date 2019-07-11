package server

import (
	"fmt"
	"os/exec"
	"runtime"
)

//Open opens web brower
func Open(url string) error {
	var cmd string
	var args []string

	fmt.Println(runtime.GOOS)

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		//protocl pre require  http https
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
