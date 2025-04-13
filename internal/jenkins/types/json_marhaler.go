package types

import (
	"encoding/json"

	"olympos.io/encoding/cjson"
)

var (
	_ json.Marshaler = (*InsecureUpdateJSON)(nil)
	_ json.Marshaler = (*SignedUpdateJSON)(nil)
)

func (o *InsecureUpdateJSON) MarshalJSON() ([]byte, error) {
	return cjson.Marshal(*o)
}

func (o *SignedUpdateJSON) MarshalJSON() ([]byte, error) {
	return cjson.Marshal(*o)
}
