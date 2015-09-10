package main

import "os"
import "log"
import "path/filepath"
import "ldpserver/web"
import "fmt"

func main() {
	rootFolder, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal("Could not determine root folder")
	}
	address := "localhost:9001"
	dataPath := rootFolder + "/data"
	numArgs := len(os.Args)

	if numArgs > 1 {
		if os.Args[1] == "--help" {
			showHelp()
			return
		}

		address = os.Args[1]
		if numArgs > 2 {
			dataPath = os.Args[2]
		}
	}

	web.Start(address, dataPath)
}

func showHelp() {
	fmt.Printf("Syntax:\n")
	fmt.Printf("    %s [address] [dataPath]\n", os.Args[0])
	fmt.Printf("\n")
	fmt.Printf("address   Represents that address where the server will listen for connections\n")
	fmt.Printf("          Defaults to localhost:9001\n")
	fmt.Printf("\n")
	fmt.Printf("dataPath  Represents the path on disk where data will be saved\n")
	fmt.Printf("          Defaults to ./data\n")
	fmt.Printf("\n")
}
