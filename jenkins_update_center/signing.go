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
	"jenkins-resigner-service/jenkins_update_center/json_schema"
	"strings"
)

var (
	signingInfo SigningInfoT
)

func ParseSigningParameters(caPath, certPath, privPath, privEncPassword string) error {
	var (
		err error
		ok  bool

		pemBytes []byte
		pemBlock *pem.Block

		cert   *x509.Certificate
		pkeyIf interface{}
	)

	if caPath != "" {
		pemBytes, err = ioutil.ReadFile(caPath)
		if err != nil {
			return fmt.Errorf("cannot load CA certificates from %s: %s", caPath, err)
		}
		log.Debug("CA certificates imported from ", caPath)

		signingInfo.roots.AppendCertsFromPEM(pemBytes)
	}

	if certPath != "" {
		pemBytes, err = ioutil.ReadFile(certPath)
		if err != nil {
			return fmt.Errorf("cannot load certificates from %s: %s", certPath, err)
		}

		pemBlock, _ = pem.Decode(pemBytes)

		if pemBlock == nil {
			return fmt.Errorf("failed to parse certificate PEM")
		}
		signingInfo.cert, err = x509.ParseCertificate(pemBlock.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse certificate: " + err.Error())
		}

		log.Debugf("Certificate loaded from %s, validity between %s and %s for ", Opts.SignCertificatePath, cert.NotBefore, cert.NotAfter)
	} else {
		return fmt.Errorf("certificate path is not provided")
	}

	if privPath != "" {
		pemBytes, err = ioutil.ReadFile(privPath)
		if err != nil {
			return fmt.Errorf("cannot load certificates from %s: %s", privPath, err)
		}
		pemBlock, _ = pem.Decode(pemBytes)
		//log.Debugf("%s, %s", pemBlock.Type, pemBlock.Headers)
		if privEncPassword != "" {
			pemBytes, err = x509.DecryptPEMBlock(pemBlock, []byte(privEncPassword))
		} else {
			pemBytes = pemBlock.Bytes
		}

		if pkeyIf, err = x509.ParsePKCS1PrivateKey(pemBytes); err != nil {
			if pkeyIf, err = x509.ParsePKCS8PrivateKey(pemBytes); err != nil { // note this returns type `interface{}`
				return fmt.Errorf("cannot load private key from %s: %s", privPath, err)
			}
		}

		signingInfo.priv, ok = pkeyIf.(*rsa.PrivateKey)
		if !ok {
			return fmt.Errorf("cannot load private key from %s: %s", privEncPassword, err)
		}

		//log.Debugf("Loaded private key with '%d' public part", pkey.Public())
	} else {
		return fmt.Errorf("private key path is not provided")
	}

	signingInfo.set = true

	return nil
}

//func (uj *UpdateJSONT) SetSigningData(roots *x509.CertPool, cert *x509.Certificate, priv *rsa.PrivateKey) error {
//	uj.mu.Lock()
//	defer func() {
//		uj.mu.Unlock()
//	}()
//
//	uj.signingInfo = SigningInfoT{
//		roots,
//		cert,
//		priv,
//		true,
//	}
//
//	return nil
//}

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

func VerifySignature(jsonData []byte, certificates []x509.Certificate, digest1, signature1, digest512, signature512 string) bool {
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

	if len(certificates) < 1 {
		log.Warn("cannot verify signature: certificates are not present")
		return false
	}
	crt = certificates[0].PublicKey.(*rsa.PublicKey)

	// SHA512...
	shaXDigest = getDigestSHA512(jsonData)
	if !isDigestsMatch(shaXDigest, digest512) {
		log.Warn("Provided and computed SHA512 digests are different")
		return false
	}
	log.Debug("SHA512 digests match")


	sig, err = hex.DecodeString(signature512)
	err = rsa.VerifyPKCS1v15(crt, crypto.SHA512, shaXDigest, sig)
	if err != nil {
		log.Warn(err)
		return false
	}
	log.Debugf("RSAWithSHA512 signature valid")

	// SHA1...
	shaXDigest = getDigestSHA1(jsonData)

	if !isDigestsMatch(shaXDigest, digest1) {
		log.Warnf("Provided and computed SHA1 digests are different: %s vs %s", digest1, base64.StdEncoding.EncodeToString(shaXDigest))
		return false
	}
	log.Debug("SHA1 digests match")

	sig, err = base64.StdEncoding.DecodeString(signature1)
	err = rsa.VerifyPKCS1v15(crt, crypto.SHA1, shaXDigest, sig)
	if err != nil {
		log.Warn(err)
		return false
	}
	log.Debug("RSAWithSHA1 signature valid")

	//fmt.Print(string(unsigned_json))
	return true
}

func (uj *UpdateJSONT) NewSignature(sInfo JSONSignatureComponents, roots *x509.CertPool, certs *x509.Certificate) (*json_schema.Signature, error) {
	return &json_schema.Signature{
		CorrectDigest512: sInfo.GetDigest512(),
		CorrectSignature512: sInfo.GetSignature512(),
		CorrectDigest: sInfo.GetDigest1(),
		CorrectSignature: sInfo.GetSignature1(),
		Certificates: []string{
			sInfo.GetCertificate(uj.signingInfo.cert),
		},
	}, nil
}

func (uj *UpdateJSONT) VerifySignature() (bool, error) {
	uj.mu.RLock()
	defer func() {
		uj.mu.RUnlock()
	}()

	rawData, err := uj.GetUnsignedJSON()
	if err != nil {
		return false, fmt.Errorf("cannot get unsigned JSON data: %s", err)
	}

	certs, err := uj.GetCertificates()
	if err != nil {
		return false, err
	}

	return VerifySignature(
		rawData,
		certs,
		uj.Signature.CorrectDigest,
		uj.Signature.CorrectSignature,
		uj.Signature.CorrectDigest512,
		uj.Signature.CorrectSignature512,
	), nil
}

func (uj *UpdateJSONT) SignPatchedJSON() error {
	if !uj.isPatched {
		return fmt.Errorf("JSON is not patched yes")
	}

	if !uj.signingInfo.set {
		return fmt.Errorf("signing info is not configured yes")
	}

	rawData, err := uj.GetUnsignedJSON()
	if err != nil {
		return fmt.Errorf("cannot get unsigned JSON data: %s", err)
	}

	sInfo, err := SignJSON(rawData, uj.signingInfo.priv)
	sign, err := uj.NewSignature(sInfo, uj.signingInfo.roots, uj.signingInfo.cert)
	if err != nil {
		return fmt.Errorf("cannot form new Signature object: %s", err)
	}

	uj.mu.Lock()
	defer func() {
		uj.mu.Unlock()
	}()

	uj.json.Signature = *sign
	uj.Signature = sign

	return nil
}

func SignJSON(jsonData []byte, priv *rsa.PrivateKey) (sInfo JSONSignatureComponents, err error) {
	sInfo.digest1 = getDigestSHA1(jsonData)
	sInfo.digest512 = getDigestSHA512(jsonData)


	sInfo.signature512, err = rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA512, sInfo.digest512)
	if err != nil {
		return sInfo, fmt.Errorf("cannot sign JSON document with SHA512WithRSA: %s", err)
	}

	sInfo.signature1, err = rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA1, sInfo.digest1)
	if err != nil {
		return sInfo, fmt.Errorf("cannot sign JSON document with SHA1WithRSA: %s", err)
	}

	return sInfo, nil
}
