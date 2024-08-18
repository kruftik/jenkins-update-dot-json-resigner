package types

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

func (s Signature) GetCertificates() ([]*x509.Certificate, error) {
	var (
		err      error
		crtBytes []byte
	)

	certs := make([]*x509.Certificate, 0, len(s.Certificates))

	for _, crtBase64 := range s.Certificates {
		crtBytes, err = base64.StdEncoding.DecodeString(crtBase64)
		if err != nil {
			return nil, fmt.Errorf("cannot decode '%s' as base64: %w", crtBase64, err)
		}

		crt, err := x509.ParseCertificate(crtBytes)
		if err != nil {
			return nil, fmt.Errorf("cannot parse '%s' as x509 cert: %w", crtBase64, err)
		}

		certs = append(certs, crt)
	}

	return certs, nil
}
