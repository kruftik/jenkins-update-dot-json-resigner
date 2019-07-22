package main

import (
	//"net/http"

	//"time"

	"fmt"
	"github.com/jessevdk/go-flags"
	"go.uber.org/zap"
	"jenkins-resigner-service/jenkins_update_center"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	GitCommit = "0.0.1"
	UpdateCenterDotJSON = "/update-center.json"
)

var (
	logger *zap.Logger
	log    *zap.SugaredLogger
	// Opts with all cli commands and flags

	Opts = struct {
		Dbg bool `long:"debug" env:"DEBUG" description:"debug mode"`

		UpdateJSONPath string `long:"update-json-path"  env:"UPDATE_JSON_PATH"`
		UpdateJSONURL  string `long:"update-json-url" env:"UPDATE_JSON_URL"`

		UpdateJSONCacheTTL time.Duration `long:"cache-ttl" env:"UPDATE_JSON_CACHE_TTL" default:"30m"`

		OriginDownloadURI string `long:"origin-download-uri" env:"ORIGIN_DOWNLOAD_URL" default:"http://updates.jenkins-ci.org/"`
		NewDownloadURI    string `long:"new-download-uri" env:"NEW_DOWNLOAD_URI" required:"true"`

		SignCAPath          string `long:"ca-certificate-path" env:"SIGN_CA_PATH" description:"x509 CA certificates path"`
		SignCertificatePath string `long:"certificate-path" env:"SIGN_CERTIFICATE_PATH" description:"x509-certificate path" required:"true"`
		SignKeyPath         string `long:"key-path" env:"SIGN_KEY_PATH" description:"private key path" required:"true"`
		SignKeyPassword     string `long:"private-key-pass" env:"SIGN_KEY_PASSWORD"`

		ServerPort			int `long:"listen-port" env:"LISTEN_PORT" default:"8282"`
	}{}

	//updateJSON *UpdateJSONT
)

func main() {
	_, err := flags.Parse(&Opts)
	if err != nil {
		fmt.Println("Can't parse flags: ", err)
		os.Exit(1)
	}

	// Logging...
	if Opts.Dbg {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	defer func() {
		_ = logger.Sync()
	}()

	zap.ReplaceGlobals(logger)
	log = zap.S()

	// Shutting down handling...
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Infow("ResignerService shutting down")
		os.Exit(0)
	}()

	log.Infof("Jenkins update.json ResignerService (v%s) starting up...", GitCommit)

	jenkins_update_center.Init()


	err = initialize()
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	//jsonMap := updateJSON.json.Value().(update_json_schema.UpdateJSON)
	//if !ok {
	//	panic("cannot unmarshal JSON")
	//}
	//v := reflect.ValueOf(updateJSON.json.Signature)
	//typeOfS := v.Type()
	////keys := v.MapKeys()
	////strkeys := make([]string, len(keys))
	//for i := 0; i < v.NumField(); i++ {
	//	fmt.Printf("%s: %s\n\n", typeOfS.Field(i).Name, v.Field(i).String())
	//}
	//insecureUpdateJSON :=

	//
	//if err != nil {
	//	log.Error(err)
	//	return
	//}

	//fmt.Print(insecureUpdateJSON.Signature.CorrectSignature512)

	//if err = updateJSON.LoadCertificates(); err != nil {
	//	log.Error(err)
	//}

	bDigestsMatch, err := updateJSON.VerifySignature()
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("isDigestsMatch: %t", bDigestsMatch)

	err = updateJSON.PatchUpdateCenterURLs()
	if err != nil {
		log.Error(err)
		return
	}
	log.Debug("JSON patched")

	err = updateJSON.SignPatchedJSON()
	if err != nil {
		log.Error(err)
		return
	}
	log.Debug("JSON resigned")

	bDigestsMatch, err = updateJSON.VerifySignature()
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("isDigestsMatch: %t", bDigestsMatch)

	err = updateJSON.SaveJSONP("jsons/resigned.json", false)
	if err != nil {
		log.Error(err)
		return
	}

}
