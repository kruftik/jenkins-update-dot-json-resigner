package json

import (
	"olympos.io/encoding/cjson"
)

func MarshalJSON[T any](ptr *T) ([]byte, error) {
	return cjson.Marshal(ptr)
}
