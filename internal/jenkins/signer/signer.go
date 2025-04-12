package signer

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/config"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/types"
)

var (
	_ types.Signer = (*Service)(nil)
)

type Service struct {
	log   *zap.SugaredLogger
	roots x509.CertPool
	cert  *x509.Certificate
	priv  *rsa.PrivateKey
}

func NewSignerService(log *zap.SugaredLogger, cfg config.SignerConfig) (*Service, error) {
	s := &Service{
		log: log,
	}

	if err := s.parseSignerParameters(cfg); err != nil {
		return nil, fmt.Errorf("signer parameters are not valid: %w", err)
	}

	return s, nil
}

func (s *Service) GetSignature(unsigned json.Marshaler) (types.Signature, error) {
	var (
		signature = JSONSignatureComponents{}
		err       error
	)

	bytez, err := unsigned.MarshalJSON()
	if err != nil {
		return types.Signature{}, fmt.Errorf("cannot marshal unsigned JSON: %w", err)
	}

	signature.digest1, signature.digest512 = getDigestSHA1(bytez), getDigestSHA512(bytez)

	signature.signature512, err = rsa.SignPKCS1v15(rand.Reader, s.priv, crypto.SHA512, signature.digest512)
	if err != nil {
		return types.Signature{}, fmt.Errorf("cannot sign JSON document with SHA512WithRSA: %w", err)
	}

	signature.signature1, err = rsa.SignPKCS1v15(rand.Reader, s.priv, crypto.SHA1, signature.digest1)
	if err != nil {
		return types.Signature{}, fmt.Errorf("cannot sign JSON document with SHA1WithRSA: %w", err)
	}

	return signature.GetSignatureObject(s.roots, s.cert), nil
}

func (s *Service) isDigestsMatch(computedDigest []byte, providedDigest string) bool {
	// SHA-512
	if strings.EqualFold(providedDigest, hex.EncodeToString(computedDigest)) {
		return true
	}

	// Base64
	if strings.EqualFold(providedDigest, base64.StdEncoding.EncodeToString(computedDigest)) {
		return true
	}

	return false
}

func (s *Service) VerifySignature(unsigned json.Marshaler, signature types.Signature) error {
	bytez, err := unsigned.MarshalJSON()
	if err != nil {
		return fmt.Errorf("cannot marshal unsigned JSON: %w", err)
	}

	certificates, err := signature.GetCertificates()
	if err != nil {
		return fmt.Errorf("cannot extract certificates from json-file: %w", err)
	}

	if len(certificates) < 1 {
		return fmt.Errorf("cannot verify signature: certificates are not present")
	}
	crt, ok := certificates[0].PublicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("cannot cast PublicKey")
	}

	// SHA512...
	shaXDigest := getDigestSHA512(bytez)
	if !s.isDigestsMatch(shaXDigest, signature.CorrectDigest512) {
		return fmt.Errorf("provided and computed SHA512 digests are different: %s vs %s", hex.EncodeToString(shaXDigest), signature.CorrectDigest512)
	}

	sig, err := hex.DecodeString(signature.CorrectSignature512)
	if err != nil {
		return fmt.Errorf("cannot decode sha512 signature: %w", err)
	}

	err = rsa.VerifyPKCS1v15(crt, crypto.SHA512, shaXDigest, sig)
	if err != nil {
		return fmt.Errorf("sha256 signature verification failed: %w", err)
	}

	// SHA1...
	shaXDigest = getDigestSHA1(bytez)

	if !s.isDigestsMatch(shaXDigest, signature.CorrectDigest) {
		return fmt.Errorf("provided and computed SHA1 digests are different: %s vs %s", hex.EncodeToString(shaXDigest), signature.CorrectDigest)
	}

	sig, err = base64.StdEncoding.DecodeString(signature.CorrectSignature)
	if err != nil {
		return fmt.Errorf("cannot base64 decode sha1 signature: %w", err)
	}

	err = rsa.VerifyPKCS1v15(crt, crypto.SHA1, shaXDigest, sig)
	if err != nil {
		return fmt.Errorf("sha1 signature verification failed: %w", err)
	}

	return nil
}
