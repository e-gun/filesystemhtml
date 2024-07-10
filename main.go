package filesystemhtml

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// as a package: start filesystemhtml.WatchFS() and then query filesystemhtml.FSResponse as needed

func main() {
	const (
		HELP = `
-h      this help
-d      directory to watch`
	)
	args := os.Args[1:len(os.Args)]

	FSDir = "./"
	for i, a := range args {
		switch a {
		case "-h":
			fmt.Printf(HELP)
			os.Exit(0)
		case "-d":
			FSDir = args[i+1]
		}
	}

	abs, err := filepath.Abs(FSDir)
	if err != nil {
		log.Fatalln(err)
	}
	AbsPath = abs + "/"

	go WatchFS()
	time.Sleep(100 * time.Millisecond)
	fmt.Println("html")
	fmt.Println(FSResponse[0])
	fmt.Println("js")
	fmt.Println(FSResponse[1])
}
