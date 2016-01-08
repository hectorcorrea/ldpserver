package main

import (
	"flag"
	"ldpserver/web"
	"os"
	"path/filepath"
)

func main() {
	rootFolder, err := filepath.Abs(filepath.Dir(os.Args[0]) + "/data")
	if err != nil {
		panic("Could not determine root folder")
	}

	var address = flag.String("address", "localhost:9001", "Address where server will listen for connections")
	var dataPath = flag.String("data", rootFolder, "Path where data will be saved")
	flag.Parse()

	web.Start(*address, *dataPath)
}
