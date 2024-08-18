package signer

import (
	"crypto/sha1"
	"crypto/sha512"
)

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
