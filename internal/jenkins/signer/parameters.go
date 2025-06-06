package signer

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/config"
)

func (s *Service) parseSignerParameters(cfg config.SignerConfig) error {
	var err error

	if s.roots, err = s.parseCACertificates(cfg.CAPath); err != nil {
		return fmt.Errorf("cannot parse CA certificates: %w", err)
	}

	if s.cert, err = s.parseCertificate(cfg.CertificatePath); err != nil {
		return fmt.Errorf("cannot parse certificate: %w", err)
	}

	if s.priv, err = s.parsePrivateKey(cfg.KeyPath, cfg.KeyPassword); err != nil {
		return fmt.Errorf("cannot parse private key: %w", err)

	}

	return nil
}

func (s *Service) parseCertificate(certPath string) (*x509.Certificate, error) {
	if certPath == "" {
		return nil, fmt.Errorf("certificate path is not provided")
	}

	pemBytes, err := os.ReadFile(certPath) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("cannot load certificates from %s: %w", certPath, err)
	}

	pemBlock, _ := pem.Decode(pemBytes)

	if pemBlock == nil {
		return nil, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	s.log.Infof("Certificate loaded from %s, validity between %s and %s for ", certPath, cert.NotBefore, cert.NotAfter)

	return cert, nil
}

func (s *Service) parseCACertificates(caPath string) (x509.CertPool, error) {
	var roots x509.CertPool

	if caPath != "" {
		pemBytes, err := os.ReadFile(caPath) //nolint:gosec
		if err != nil {
			return x509.CertPool{}, fmt.Errorf("cannot load CA certificates from %s: %w", caPath, err)
		}

		roots.AppendCertsFromPEM(pemBytes)

		s.log.Info("CA certificates imported from ", caPath)
	}

	return roots, nil
}

func (s *Service) parsePrivateKey(privPath, privEncPassword string) (*rsa.PrivateKey, error) {
	if privPath == "" {
		return nil, fmt.Errorf("private key path is not provided")
	}

	pemBytes, err := os.ReadFile(privPath) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("cannot load certificates from %s: %w", privPath, err)
	}
	pemBlock, _ := pem.Decode(pemBytes)

	if privEncPassword != "" {
		//nolint:staticcheck
		pemBytes, err = x509.DecryptPEMBlock(pemBlock, []byte(privEncPassword))
		if err != nil {
			return nil, err
		}
	} else {
		pemBytes = pemBlock.Bytes
	}

	var (
		pkeyIf interface{}
	)

	if pkeyIf, err = x509.ParsePKCS1PrivateKey(pemBytes); err != nil {
		if pkeyIf, err = x509.ParsePKCS8PrivateKey(pemBytes); err != nil { // note this returns type `interface{}`
			return nil, fmt.Errorf("cannot load private key from %s: %w", privPath, err)
		}
	}

	priv, ok := pkeyIf.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("cannot cast private key: %w", err)
	}

	s.log.Infof("Loaded private key loaded from %s", privPath)

	return priv, nil
}
