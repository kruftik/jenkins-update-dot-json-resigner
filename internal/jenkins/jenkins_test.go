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

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	signer, err := signer.NewSignerService(log, config.SignerConfig{
		CertificatePath: "../../testdata/certs/test.crt",
		KeyPath:         "../../testdata/certs/test.key",
	})
	if err != nil {
		t.Fatal(err)
	}

	source := "https://updates.jenkins.io/current/update-center.json"
	p, err := remoteurl.NewRemoteURLProvider(log, source)
	if err != nil {
		t.Fatal(err)
	}

	juc := NewJenkinsUpdateCenter(log, config.AppConfig{
		DataDirPath:              "/tmp",
		GetUpdateJSONBodyTimeout: 128 * time.Second,
	}, p, signer, nil)

	if err := juc.RefreshContent(ctx); err != nil {
		t.Fatal(err)
	}

	t.Logf("%s is still supported", source)
}
