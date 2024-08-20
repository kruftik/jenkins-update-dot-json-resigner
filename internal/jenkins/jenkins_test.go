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

	var originalFile bytes.Buffer

	if _, err := io.Copy(&originalFile, r); err != nil {
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

	bytez, err := json.MarshalJSON(signedJSON)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(originalFile.Bytes(), bytez) {
		t.Fatalf("original and re-marshaled files do not match")
	}
}
