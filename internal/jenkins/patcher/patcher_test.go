package patcher

import (
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/config"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/types"
)

func TestPatcher(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	originURL := "http://origin.local/download/"
	patchedURL := "http://patched.local/jenkins/"

	origin := &types.InsecureUpdateJSON{
		Plugins: map[string]types.Plugin{
			"text": {
				URL: originURL + "/plugin.hpi",
			},
		},
		UpdateCenterVersion: "123",
	}

	p := NewPatcher(logger.Sugar(), config.PatchConfig{
		OriginDownloadURL: originURL,
		NewDownloadURL:    patchedURL,
	})

	if err := p.Patch(origin); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(origin.Plugins["text"].URL, originURL) {
		t.Fatal("plugin URL contain origin URL but not patched one")
	}

	if !strings.Contains(origin.Plugins["text"].URL, patchedURL) {
		t.Fatal("plugin URL does not contain patched URL")
	}
}
