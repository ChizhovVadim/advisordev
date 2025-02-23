package main

import (
	"advisordev/internal/cli"
	"log"
)

func main() {
	var app = &cli.App{}
	app.AddCommand("status", statusHandler)
	app.AddCommand("report", reportHandler)
	app.AddCommand("testdownload", testDownloadHandler)
	app.AddCommand("update", updateHandler)
	var err = app.Run()
	if err != nil {
		log.Println(err)
		return
	}
}
