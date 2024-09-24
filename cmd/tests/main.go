package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	var err = run(os.Args[1:])
	if err != nil {
		log.Println(err)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("command not specified")
	}
	var cmdName = args[0]
	args = args[1:]
	switch cmdName {
	case "history":
		return historyHandler(args)
	case "status":
		return statusHandler(args)
	case "update":
		return updateHandler(args)
	case "testdownload":
		return testDownloadHandler(args)
	default:
		return fmt.Errorf("bad command %v", cmdName)
	}
}
