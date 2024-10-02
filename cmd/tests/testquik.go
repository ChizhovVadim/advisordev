package main

import (
	"advisordev/internal/quik"
	"context"
	"flag"

	"golang.org/x/sync/errgroup"
)

func testQuikHandler(args []string) error {
	var (
		port int = 34130
	)

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.IntVar(&port, "port", port, "")
	flagset.Parse(args)

	mainConn, err := quik.InitConnection(port)
	if err != nil {
		return err
	}
	defer mainConn.Close()

	var quikService = quik.NewQuikService(mainConn)

	callbackConn, err := quik.InitConnection(port + 1)
	if err != nil {
		return err
	}
	defer callbackConn.Close()

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		for _, _ = range quik.QuikCallbacks(callbackConn) {
		}
		return nil
	})

	g.Go(func() error {
		defer callbackConn.Close()
		return quikService.MessageInfo("Привет из go")
	})

	return g.Wait()
}
