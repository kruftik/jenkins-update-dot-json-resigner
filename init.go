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

	juc, err = jenkins_update_center.NewJenkinsUC(Opts.UpdateJSONURL, Opts.UpdateJSONPath, Opts.UpdateJSONCacheTTL)
	if err != nil {
		return err
	}

	err = initHTTP(juc)
	if err != nil {
		return err
	}

	return nil
}
