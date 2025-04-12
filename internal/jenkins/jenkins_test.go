package jenkins

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/config"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/signer"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/sourcefileproviders/localfile"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/sourcefileproviders/remoteurl"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/json"
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

func TestSignedUpdateJSON_MarshalJSON(t *testing.T) {
	var (
		logger, _ = zap.NewDevelopment()
		log       = logger.Sugar()
	)

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	p, err := localfile.NewLocalFileProvider("../../testdata/update-center/update-center.jsonp")
	if err != nil {
		t.Fatal(err)
	}

	_, r, err := p.GetBody(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	originalFileBytes, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	juc := NewJenkinsUpdateCenter(log, config.AppConfig{
		DataDirPath:              "/tmp",
		GetUpdateJSONBodyTimeout: 3 * time.Second,
	}, p, nil, nil)

	_, signedJSON, err := juc.getOriginal(ctx)
	if err != nil {
		t.Fatal(err)
	}

	marshaledFileBytes, err := json.MarshalJSON(signedJSON)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(originalFileBytes, marshaledFileBytes) {
		t.Fatalf("original and re-marshaled files do not match")
	}
}
