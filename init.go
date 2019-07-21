package main

import (
//"fmt"
//"os"
//"os/signal"
//"syscall"
//
//"github.com/jessevdk/go-flags"
//"go.uber.org/zap"
)

func initialize() error {
	roots, cert, priv, err := parseSigningParameters()
	if err != nil {
		return err
	}

	updateJSON, err = parseUpdateJSONLocation()
	if err != nil {
		return err
	}

	err = updateJSON.SetSigningData(roots, cert, priv)
	if err != nil {
		return err
	}

	return nil
}
