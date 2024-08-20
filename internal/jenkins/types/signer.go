package types

type Signer interface {
	GetSignature(unsinged *InsecureUpdateJSON) (Signature, error)
	VerifySignature(unsinged *InsecureUpdateJSON, signature Signature) error
}
