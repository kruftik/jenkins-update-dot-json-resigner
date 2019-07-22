package main

import (
//"fmt"
//"os"
//"os/signal"
//"syscall"
//
//"github.com/jessevdk/go-flags"
//"go.uber.org/zap"
	"jenkins-resigner-service/jenkins_update_center"
)

func initialize() error {
	err := jenkins_update_center.ParseSigningParameters(
		Opts.SignCertificatePath,
		Opts.SignCertificatePath,
		Opts.SignKeyPath,
		Opts.SignKeyPassword,
	)
	if err != nil {
		return err
	}

	err = jenkins_update_center.ParseUpdateJSONLocation(Opts.UpdateJSONURL, Opts.UpdateJSONPath)
	if err != nil {
		return err
	}

	err = initHTTP()
	if err != nil {
		return err
	}

	return nil
}
