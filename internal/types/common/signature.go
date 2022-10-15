package common

type SignatureV1 struct {
	Certificates []string `json:"certificates"`

	CorrectDigest    string `json:"correct_digest"`
	CorrectDigest512 string `json:"correct_digest512"`

	CorrectSignature    string `json:"correct_signature"`
	CorrectSignature512 string `json:"correct_signature512"`

	Digest       string `json:"digest"`
	Digest512    string `json:"digest512"`
	Signature    string `json:"signature"`
	Signature512 string `json:"signature512"`
}

type SignatureV2 struct {
	Certificates []string `json:"certificates"`

	CorrectDigest    string `json:"correct_digest"`
	CorrectDigest512 string `json:"correct_digest512"`

	CorrectSignature    string `json:"correct_signature"`
	CorrectSignature512 string `json:"correct_signature512"`

	// incorrect digest and signatures are not included anymore
	//Digest              string   `json:"digest"`
	//Digest512           string   `json:"digest512"`
	//Signature           string   `json:"signature"`
	//Signature512        string   `json:"signature512"`
}
