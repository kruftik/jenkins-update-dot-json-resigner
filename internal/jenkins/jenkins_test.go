package jenkins

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/config"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/signer"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/sourcefileproviders/remoteurl"
)

func TestCurrentUpdateJSON(t *testing.T) {
	var (
		logger, _ = zap.NewDevelopment()
		log       = logger.Sugar()
	)

	source := "https://updates.jenkins.io/current/update-center.json"

	signer, err := signer.NewSignerService(log, config.SignerConfig{
		CertificatePath: "../../testdata/certs/test.crt",
		KeyPath:         "../../testdata/certs/test.key",
	})
	if err != nil {
		t.Fatal(err)
	}

	p, err := remoteurl.NewRemoteURLProvider(log, source)
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewJenkinsUpdateCenter(context.Background(), log, config.AppConfig{
		DataDirPath:               "/tmp",
		UpdateJSONDownloadTimeout: 128 * time.Second,
	}, p, signer, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s is still supported", source)
}
