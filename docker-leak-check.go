package main

import (
	"flag"
	"log"
	"runtime"

	"github.com/voleer/docker-leak-check/pkg"
)

func main() {
	var folder string
	defaultFolder := `C:\ProgramData\docker`
	if runtime.GOOS != "windows" {
		defaultFolder = `/var/lib/docker`
	}
	var remove bool
	flag.StringVar(&folder, "folder", defaultFolder, "Root of the Docker runtime")
	flag.BoolVar(&remove, "remove", false, "Remove unreferenced layers")
	flag.Parse()

	if err := pkg.Run(folder, remove); err != nil {
		log.Fatalln(err)
	}
}
