package types

import (
	"olympos.io/encoding/cjson"
)

func (o *InsecureUpdateJSON) MarshalJSON() ([]byte, error) {
	return cjson.Marshal(*o)
}

func (o *SignedUpdateJSON) MarshalJSON() ([]byte, error) {
	return cjson.Marshal(*o)
}
