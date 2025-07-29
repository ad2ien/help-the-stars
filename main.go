package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var version = "1.0.0"

func main() {
	helpFlag := flag.Bool("help", false, "Display help information")
	versionFlag := flag.Bool("version", false, "Display version information")

	flag.Parse()

	if *helpFlag {
		fmt.Println("Usage of Help the stars:")
		fmt.Println("  -help\tDisplay help information")
		fmt.Println("  -version\tDisplay version information")
		os.Exit(0)
	}

	if *versionFlag {
		fmt.Println("Version:", version)
		os.Exit(0)
	}

	GetSettings()
	res, err := GetStaredRepos(50)
	if err != nil {
		log.Fatal("Gh Request error : ", err)
	}

	log.Println(res)

}
