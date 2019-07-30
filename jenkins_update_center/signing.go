package jenkins_update_center

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"strings"
)

type SigningInfoT struct {
	roots *x509.CertPool
	cert  *x509.Certificate
	priv  *rsa.PrivateKey
	set   bool
}

type JSONSignatureComponents struct {
	digest1    []byte
	signature1 []byte

	digest512    []byte
	signature512 []byte
}

var (
	signingInfo SigningInfoT
)

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

func (sc JSONSignatureComponents) GetCertificates(roots *x509.CertPool, cert *x509.Certificate) []string {
	return []string{base64.StdEncoding.EncodeToString(cert.Raw)}
}

func (sc JSONSignatureComponents) GetSignatureObject(roots *x509.CertPool, cert *x509.Certificate) *Signature {
	return &Signature{
		Certificates:        sc.GetCertificates(roots, cert),
		CorrectDigest:       sc.GetDigest1(),
		CorrectDigest512:    sc.GetDigest512(),
		CorrectSignature:    sc.GetSignature1(),
		CorrectSignature512: sc.GetSignature512(),
		Digest:              "",
		Digest512:           "",
		Signature:           "",
		Signature512:        "",
	}
}

func ParseSigningParameters(caPath, certPath, privPath, privEncPassword string) (*SigningInfoT, error) {
	var (
		err error
		ok  bool

		pemBytes []byte
		pemBlock *pem.Block

		roots  = &x509.CertPool{}
		cert   *x509.Certificate
		pkeyIf interface{}
		priv   *rsa.PrivateKey
	)

	if caPath != "" {
		pemBytes, err = ioutil.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("cannot load CA certificates from %s: %s", caPath, err)
		}
		log.Debug("CA certificates imported from ", caPath)

		roots.AppendCertsFromPEM(pemBytes)
	}

	if certPath != "" {
		pemBytes, err = ioutil.ReadFile(certPath)
		if err != nil {
			return nil, fmt.Errorf("cannot load certificates from %s: %s", certPath, err)
		}

		pemBlock, _ = pem.Decode(pemBytes)

		if pemBlock == nil {
			return nil, fmt.Errorf("failed to parse certificate PEM")
		}
		cert, err = x509.ParseCertificate(pemBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: " + err.Error())
		}

		log.Debugf("Certificate loaded from %s, validity between %s and %s for ", certPath, cert.NotBefore, cert.NotAfter)
	} else {
		return nil, fmt.Errorf("certificate path is not provided")
	}

	if privPath != "" {
		pemBytes, err = ioutil.ReadFile(privPath)
		if err != nil {
			return nil, fmt.Errorf("cannot load certificates from %s: %s", privPath, err)
		}
		pemBlock, _ = pem.Decode(pemBytes)
		//log.Debugf("%s, %s", pemBlock.Type, pemBlock.Headers)
		if privEncPassword != "" {
			pemBytes, err = x509.DecryptPEMBlock(pemBlock, []byte(privEncPassword))
			if err != nil {
				return nil, err
			}
		} else {
			pemBytes = pemBlock.Bytes
		}

		if pkeyIf, err = x509.ParsePKCS1PrivateKey(pemBytes); err != nil {
			if pkeyIf, err = x509.ParsePKCS8PrivateKey(pemBytes); err != nil { // note this returns type `interface{}`
				return nil, fmt.Errorf("cannot load private key from %s: %s", privPath, err)
			}
		}

		priv, ok = pkeyIf.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("cannot load private key from %s: %s", privEncPassword, err)
		}

		//log.Debugf("Loaded private key with '%d' public part", pkey.Public())
	} else {
		return nil, fmt.Errorf("private key path is not provided")
	}

	signingInfo = SigningInfoT{
		roots: roots,
		cert:  cert,
		priv:  priv,
		set:   true,
	}

	return &signingInfo, nil
}

func (uj UpdateJSON) GetCertificates() ([]x509.Certificate, error) {
	var (
		sign = uj.Signature

		err      error
		crtBytes []byte
	)

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

		certs[idx] = *crt
	}

	return certs, nil
}

func getDigestSHA1(data []byte) []byte {
	h := sha1.New()
	h.Write(data)
	return h.Sum(nil)
	//return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func getDigestSHA512(data []byte) []byte {
	h := sha512.New()
	h.Write(data)
	return h.Sum(nil)
	//return hex.EncodeToString(h.Sum(nil))
}

func (uj UpdateJSON) VerifySignature() error {
	var (
		err        error
		crt        *rsa.PublicKey
		shaXDigest []byte
		sig        []byte
	)

	isDigestsMatch := func(computedDigest []byte, providedDigest string) bool {
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

	log.Info("Verifying JSON document signature...")

	certificates, err := uj.GetCertificates()
	if err != nil {
		return err
	}

	if len(certificates) < 1 {
		return fmt.Errorf("cannot verify signature: certificates are not present")
	}
	crt = certificates[0].PublicKey.(*rsa.PublicKey)

	jsonData, err := getUnsignedJSON(uj)
	if err != nil {
		return err
	}

	// SHA512...
	shaXDigest = getDigestSHA512(jsonData)
	if !isDigestsMatch(shaXDigest, uj.Signature.CorrectDigest512) {
		fmt.Print(string(jsonData[:128]) + "\n")
		fmt.Print(string(jsonData[len(jsonData)-129:]) + "\n")
		return fmt.Errorf("provided and computed SHA512 digests are different")
	}
	log.Debug("SHA512 digests match")

	sig, err = hex.DecodeString(uj.Signature.CorrectSignature512)
	if err != nil {
		return err
	}

	err = rsa.VerifyPKCS1v15(crt, crypto.SHA512, shaXDigest, sig)
	if err != nil {
		return err
	}
	log.Debugf("RSAWithSHA512 signature valid")

	// SHA1...
	shaXDigest = getDigestSHA1(jsonData)

	if !isDigestsMatch(shaXDigest, uj.Signature.CorrectDigest) {
		return fmt.Errorf("provided and computed SHA1 digests are different")
	}
	log.Debug("SHA1 digests match")

	sig, err = base64.StdEncoding.DecodeString(uj.Signature.CorrectSignature)
	if err != nil {
		return err
	}

	err = rsa.VerifyPKCS1v15(crt, crypto.SHA1, shaXDigest, sig)
	if err != nil {
		return err
	}
	log.Debug("RSAWithSHA1 signature valid")

	log.Debug("Verifying JSON document signature [done]")

	//fmt.Print(string(unsigned_json))
	return nil
}

func (sInfo *SigningInfoT) SignJSONData(jsonData *InsecureUpdateJSON) (*Signature, error) {
	signature := JSONSignatureComponents{}

	bytesData, err := jsonData.GetBytes()
	if err != nil {
		return nil, err
	}

	signature.digest1, signature.digest512 = getDigestSHA1(bytesData), getDigestSHA512(bytesData)

	signature.signature512, err = rsa.SignPKCS1v15(rand.Reader, sInfo.priv, crypto.SHA512, signature.digest512)
	if err != nil {
		return nil, fmt.Errorf("cannot sign JSON document with SHA512WithRSA: %s", err)
	}

	signature.signature1, err = rsa.SignPKCS1v15(rand.Reader, sInfo.priv, crypto.SHA1, signature.digest1)
	if err != nil {
		return nil, fmt.Errorf("cannot sign JSON document with SHA1WithRSA: %s", err)
	}

	return signature.GetSignatureObject(sInfo.roots, sInfo.cert), nil
}
