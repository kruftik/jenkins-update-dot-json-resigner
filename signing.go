package main

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

func parseSigningParameters() (*x509.CertPool, *x509.Certificate, *rsa.PrivateKey, error) {
	var (
		err error
		ok  bool

		pemBytes []byte
		pemBlock *pem.Block

		ca = x509.NewCertPool()

		cert   *x509.Certificate
		pkeyIf interface{}
		pkey   *rsa.PrivateKey
	)

	if Opts.SignCAPath != "" {
		pemBytes, err = ioutil.ReadFile(Opts.SignCAPath)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("cannot load CA certificates from %s: %s", Opts.SignCAPath, err)
		}
		log.Debugf("CA certificates imported from " + Opts.SignCAPath)

		ca.AppendCertsFromPEM(pemBytes)
	}

	if Opts.SignCertificatePath != "" {
		pemBytes, err = ioutil.ReadFile(Opts.SignCertificatePath)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("cannot load certificates from %s: %s", Opts.SignCertificatePath, err)
		}

		pemBlock, _ = pem.Decode(pemBytes)

		if pemBlock == nil {
			return nil, nil, nil, fmt.Errorf("failed to parse certificate PEM")
		}
		cert, err = x509.ParseCertificate(pemBlock.Bytes)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to parse certificate: " + err.Error())
		}

		log.Debugf("Certificate loaded from %s, validity between %s and %s for ", Opts.SignCertificatePath, cert.NotBefore, cert.NotAfter)
	} else {
		return nil, nil, nil, fmt.Errorf("certificate path is not provided")
	}

	if Opts.SignKeyPath != "" {
		pemBytes, err = ioutil.ReadFile(Opts.SignKeyPath)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("cannot load certificates from %s: %s", Opts.SignCertificatePath, err)
		}
		pemBlock, _ = pem.Decode(pemBytes)
		//log.Debugf("%s, %s", pemBlock.Type, pemBlock.Headers)
		if Opts.SignKeyPassword != "" {
			pemBytes, err = x509.DecryptPEMBlock(pemBlock, []byte(Opts.SignKeyPassword))
		} else {
			pemBytes = pemBlock.Bytes
		}

		if pkeyIf, err = x509.ParsePKCS1PrivateKey(pemBytes); err != nil {
			if pkeyIf, err = x509.ParsePKCS8PrivateKey(pemBytes); err != nil { // note this returns type `interface{}`
				return nil, nil, nil, fmt.Errorf("cannot load private key from %s: %s", Opts.SignKeyPath, err)
			}
		}

		pkey, ok = pkeyIf.(*rsa.PrivateKey)
		if !ok {
			return nil, nil, nil, fmt.Errorf("cannot load private key from %s: %s", Opts.SignKeyPath, err)
		}

		//log.Debugf("Loaded private key with '%d' public part", pkey.Public())
	} else {
		return nil, nil, nil, fmt.Errorf("private key path is not provided")
	}

	return ca, cert, pkey, nil
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

	sig, err = hex.DecodeString(signature512)
	err = rsa.VerifyPKCS1v15(crt, crypto.SHA512, shaXDigest, sig)
	if err != nil {
		log.Warn(err)
	}

	// SHA1...
	shaXDigest = getDigestSHA1(jsonData)

	if !isDigestsMatch(shaXDigest, digest1) {
		log.Warnf("Provided and computed SHA1 digests are different: %s vs %s", digest1, base64.StdEncoding.EncodeToString(shaXDigest))
		return false
	}

	sig, err = base64.StdEncoding.DecodeString(signature1)
	err = rsa.VerifyPKCS1v15(crt, crypto.SHA1, shaXDigest, sig)
	if err != nil {
		log.Warn(err)
	}

	//fmt.Print(string(unsigned_json))
	return true
}

func SignJSON(jsonData []byte, priv *rsa.PrivateKey) (digest1str, signature1str, digest512str, signature512str string, err error) {
	var (
		digest1   []byte
		digest512 []byte

		signature1   []byte
		signature512 []byte

		//sig        []byte
	)

	digest512 = getDigestSHA512(jsonData)
	digest512str = base64.StdEncoding.EncodeToString(digest512)
	digest1 = getDigestSHA1(jsonData)
	digest1str = hex.EncodeToString(digest1)

	signature512, err = rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA512, digest512)
	if err != nil {
		return "", "", "", "", fmt.Errorf("cannot sign JSON document with SHA512WithRSA: %s", err)
	}
	signature512str = base64.StdEncoding.EncodeToString(signature512)

	signature1, err = rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA1, digest1)
	if err != nil {
		return "", "", "", "", fmt.Errorf("cannot sign JSON document with SHA1WithRSA: %s", err)
	}
	signature1str = base64.StdEncoding.EncodeToString(signature1)

	return digest1str, signature1str, digest512str, signature512str, nil
}
