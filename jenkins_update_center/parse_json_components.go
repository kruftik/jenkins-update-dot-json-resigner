package jenkins_update_center

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

func (uj *UpdateJSONT) GetCertificates() ([]x509.Certificate, error) {
	var (
		sign = uj.Signature

		err      error
		crtBytes []byte
	)

	uj.mu.RLock()
	defer func() {
		uj.mu.RUnlock()
	}()

	certs := make([]x509.Certificate, len(sign.Certificates))

	for idx, crtBase64 := range sign.Certificates {
		crtBytes, err = base64.StdEncoding.DecodeString(crtBase64)
		if err != nil {
			return nil, fmt.Errorf("cannot decode '%s' as base64: %s", crtBase64, err)
		}

		crt, err := x509.ParseCertificate(crtBytes)
		if err != nil {
			return nil, fmt.Errorf("cannot parse '%s' as x509 cert: %s", crtBase64, err)
		}

		log.Debugf("Cert valid before %s and %s for %s", crt.NotBefore, crt.NotAfter, crt.Subject)

		certs[idx] = *crt
	}

	return certs, nil
}
