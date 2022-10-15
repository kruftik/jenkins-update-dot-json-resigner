package main

import (
	"os"
	"sync"
	"syscall"
	"testing"
	"time"
	//log "github.com/go-pkgz/lgr"
	//"github.com/go-pkgz/repeater"
	//
	//"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/require"
)

func Test_Main(t *testing.T) {
	env := map[string]string{
		"DEBUG":                 "true",
		"SIGN_CERTIFICATE_PATH": "cert/your-update-center.crt",
		"SIGN_KEY_PATH":         "cert/your-update-center.key",
		"NEW_DOWNLOAD_URI":      "http://ftp-nyc.osuosl.org/pub/jenkins/",
		"UPDATE_JSON_PATH":      "/var/folders/w1/s4rtqcxs6ks7y61ym4gd2h780000gn/T/update.json",
//		"http_proxy":            "http://10.16.0.2:3128",
		"ORIGIN_DOWNLOAD_URL":   "http://updates.jenkins-ci.org/download/",
	}

	os.Args = os.Args[:1]

	os.Clearenv()
	for k, v := range env {
		err := os.Setenv(k, v)
		if err != nil {
			t.Error(err)
		}
	}

	go func() {
		time.Sleep(26 * time.Second)
		err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		if err != nil {
			t.Error(err)
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		//st := time.Now()
		main()
		//assert.True(t, time.Since(st).Seconds() >= 5, "should take about 5s")
		wg.Done()
	}()

	wg.Wait()
}
