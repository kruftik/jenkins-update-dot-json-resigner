package jenkins_update_center

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"jenkins-resigner-service/jenkins_update_center/json_schema"
)

func extractJSONDocument(s []byte) ([]byte, error) {
	idxFrom := bytes.Index(s, []byte(`{`))
	idxTo := bytes.LastIndex(s, []byte(`}`))

	if idxFrom == -1 || idxTo == -1 {
		return nil, fmt.Errorf("cannot find a valid JSON document in the provided string")
	}

	return s[idxFrom : idxTo+1], nil

	//sLen := len(s)

	//prefixLen := len(wrappedJSONPrefix)
	//postfixLen := len(wrappedJSONPostfix)
	//if s[:prefixLen] != wrappedJSONPrefix {
	//	return "", fmt.Errorf("given JSON-wrapped string does not begin with '%s' prefix", wrappedJSONPrefix)
	//}
	//
	//if s[sLen-postfixLen:] != wrappedJSONPostfix {
	//	return "", fmt.Errorf("given JSON-wrapped string does not end with '%s' postfix", wrappedJSONPostfix)
	//}

	//return s[prefixLen : sLen-postfixLen], nil
}

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

func (uj *UpdateJSONT) GetUnsignedJSON() ([]byte, error) {
	var insecureUpdateJSON json_schema.InsecureUpdateJSON

	uj.mu.RLock()
	defer func() {
		uj.mu.RUnlock()
	}()

	insecureUpdateJSON = json_schema.InsecureUpdateJSON(*updateJSON.json)

	data, err := json.Marshal(insecureUpdateJSON)
	if err != nil {
		return nil, err
	}

	return replaceSymbolsByTrickyMap(data), nil
}
