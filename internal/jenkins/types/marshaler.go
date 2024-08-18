package types

import (
	"encoding/json"
	"io"
)

type Marshaler interface {
	json.Marshaler

	MarshalJSONTo(w io.Writer) error
}
