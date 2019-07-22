package jenkins_update_center

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"io"
	"jenkins-resigner-service/jenkins_update_center/json_schema"
	"sync"
)

type SigningInfoT struct {
	roots *x509.CertPool
	cert  *x509.Certificate
	priv  *rsa.PrivateKey
	set   bool
}

type JSONSignatureComponents struct {
	digest1 []byte
	signature1 []byte

	digest512 []byte
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

func (sc JSONSignatureComponents) GetCertificate(cert *x509.Certificate) string {
	return base64.StdEncoding.EncodeToString(cert.Raw)
}

type jsonSymbolReplacementRuleT struct {
	from []byte
	to   []byte
}

type JenkinsUCJSONT struct {
	src string

	Get func() (io.Reader, error)
	//isRemoteSource bool // true - URL; false - file
}

type SyncedByteBuffer struct {
	mu sync.RWMutex
	data bytes.Buffer
}

func (sbb *SyncedByteBuffer) Reset() {
	sbb.mu.Lock()
	defer func() {
		sbb.mu.Unlock()
	}()

	sbb.data.Reset()
}

type UpdateJSONT struct {
	mu sync.RWMutex

	json      *json_schema.UpdateJSON
	Signature *json_schema.Signature
	//data map[string]interface{}

	signingInfo SigningInfoT

	isPatched bool
}