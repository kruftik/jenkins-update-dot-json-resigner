package signer

import (
	"crypto/sha1" //nolint:gosec
	"crypto/sha512"
)

func getDigestSHA1(data []byte) []byte {
	h := sha1.New() //nolint:gosec
	h.Write(data)
	return h.Sum(nil)
}

func getDigestSHA512(data []byte) []byte {
	h := sha512.New()
	h.Write(data)
	return h.Sum(nil)
}
