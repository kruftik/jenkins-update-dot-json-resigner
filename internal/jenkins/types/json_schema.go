package types

import (
	"fmt"
)

type InsecureUpdateJSON struct {
	ConnectionCheckURL  string                 `json:"connectionCheckUrl"`
	Core                Core                   `json:"core"`
	Deprecations        map[string]interface{} `json:"deprecations"`
	GenerationTimestamp string                 `json:"generationTimestamp"`
	ID                  string                 `json:"id"`
	Plugins             Plugins                `json:"plugins"`
	UpdateCenterVersion string                 `json:"updateCenterVersion"`
	Warnings            []interface{}          `json:"warnings"`
}

type SignedUpdateJSON struct {
	*InsecureUpdateJSON
	Signature Signature `json:"signature"`
}

func (o *SignedUpdateJSON) Sign(signer Signer) error {
	signature, err := signer.GetSignature(o.GetUnsigned())
	if err != nil {
		return fmt.Errorf("cannot calculate signature: %w", err)
	}

	if err := signer.VerifySignature(o.InsecureUpdateJSON, signature); err != nil {
		return fmt.Errorf("cannot verify signature: %w", err)
	}

	o.Signature = signature

	return nil
}

func (o *SignedUpdateJSON) GetUnsigned() *InsecureUpdateJSON {
	return o.InsecureUpdateJSON
}
