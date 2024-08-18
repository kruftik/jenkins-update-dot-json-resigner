package types

type Signer interface {
	GetSignature(unsinged Marshaler) (Signature, error)
	VerifySignature(unsinged Marshaler, signature Signature) error
}
