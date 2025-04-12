package types_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/config"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/sourcefileproviders/localfile"
)

func TestMarshalJSON(t *testing.T) {
	var (
		logger, _ = zap.NewDevelopment()
		log       = logger.Sugar()
	)

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	p, err := localfile.NewLocalFileProvider("../../../testdata/update-center/update-center.jsonp")
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

	juc := jenkins.NewJenkinsUpdateCenter(log, config.AppConfig{
		DataDirPath:              "/tmp",
		GetUpdateJSONBodyTimeout: 3 * time.Second,
	}, p, nil, nil)

	_, signedJSON, err := juc.GetOriginal(ctx)
	if err != nil {
		t.Fatal(err)
	}

	marshaledFileBytes, err := signedJSON.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(originalFileBytes, marshaledFileBytes) {
		t.Fatalf("original and re-marshaled files do not match")
	}
}
