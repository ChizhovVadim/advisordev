package main

import (
	"advisordev/internal/quik"
	"context"
	"errors"
	"flag"
	"fmt"

	"golang.org/x/sync/errgroup"
)

func testQuikHandler(args []string) error {
	var (
		port int = 34132
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

		isConn, err := quikService.IsConnected()
		if err != nil {
			return err
		}
		if !isConn {
			return errors.New("quik is not connected")
		}

		err = quikService.MessageInfo("Привет из go")
		if err != nil {
			return err
		}

		const (
			ClassCode = "SPBFUT"
			SecCode   = "CRZ4"
		)

		{
			secInfo, err := quik.GetSecurityInfo(quikService, ClassCode, SecCode)
			if err != nil {
				return err
			}
			fmt.Printf("%#v\n", secInfo)
			fmt.Println(quik.AsInt(secInfo["scale"]))
			fmt.Println(quik.AsInt(secInfo["lot_size"]))
			fmt.Println(quik.AsFloat64(secInfo["min_price_step"]))
		}

		{
			paramPriceLast, err := quik.GetParamEx(quikService, ClassCode, SecCode, quik.ParamNameLAST)
			if err != nil {
				return err
			}
			fmt.Printf("%#v\n", paramPriceLast)
			fmt.Println(quik.AsFloat64(paramPriceLast["param_value"]))
		}

		return nil
	})

	return g.Wait()
}
