package signer

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/types"
)

type JSONSignatureComponents struct {
	digest1    []byte
	signature1 []byte

	digest512    []byte
	signature512 []byte
}

func (sc JSONSignatureComponents) GetDigest512() string {
	return base64.StdEncoding.EncodeToString(sc.digest512)
}

func (sc JSONSignatureComponents) GetSignature512() string {
	return hex.EncodeToString(sc.signature512)
}

func (sc JSONSignatureComponents) GetDigest1() string {
	return hex.EncodeToString(sc.digest1)
}

func (sc JSONSignatureComponents) GetSignature1() string {
	return base64.StdEncoding.EncodeToString(sc.signature1)
}

func (sc JSONSignatureComponents) GetCertificates(_ x509.CertPool, cert *x509.Certificate) []string {
	return []string{base64.StdEncoding.EncodeToString(cert.Raw)}
}

func (sc JSONSignatureComponents) GetSignatureObject(roots x509.CertPool, cert *x509.Certificate) types.Signature {
	return types.Signature{
		Certificates:        sc.GetCertificates(roots, cert),
		CorrectDigest:       sc.GetDigest1(),
		CorrectDigest512:    sc.GetDigest512(),
		CorrectSignature:    sc.GetSignature1(),
		CorrectSignature512: sc.GetSignature512(),
	}
}
