package main

import (
	"flag"
	"runtime"

	"voleer.io/docker-leak-check/pkg"
)

func main() {
	var folder string
	defaultFolder := `C:\ProgramData\docker`
	if runtime.GOOS != "windows" {
		defaultFolder = `/var/lib/docker`
	}
	var remove bool
	flag.StringVar(&folder, "folder", "", "Root of the Docker runtime (default \""+defaultFolder+"\")")
	flag.BoolVar(&remove, "remove", false, "Remove unreferenced layers")
	flag.Parse()
	if folder == "" {
		folder = defaultFolder
	}

	pkg.Run(folder, remove)
}
