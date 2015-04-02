package main

import "os"
import "log"
import "path/filepath"
import "ldpserver/web"

func main() {
	rootFolder, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal("Could not determine root folder")
	}
	address := "localhost:9001"
	dataPath := rootFolder + "/data"

	if len(os.Args) == 2 {
		dataPath = os.Args[1]
	}

	web.Start(address, dataPath)
}
