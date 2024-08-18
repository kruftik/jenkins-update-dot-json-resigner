package types

import (
	"encoding/json"
)

type Signer interface {
	GetSignature(unsinged json.Marshaler) (Signature, error)
	VerifySignature(unsinged json.Marshaler, signature Signature) error
}
