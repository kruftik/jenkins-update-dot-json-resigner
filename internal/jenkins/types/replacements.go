package types

import (
	"bytes"
)

type jsonSymbolReplacementRuleT struct {
	from []byte
	to   []byte
}

var (
	jsonSymbolReplacementsMap = []jsonSymbolReplacementRuleT{
		{[]byte("\\u0026"), []byte("&")},
		{[]byte("\\u003c/"), []byte("<\\/")},
		{[]byte("\\u003c"), []byte("<")},
		{[]byte("\\u003e"), []byte(">")},
	}
)

func replaceSymbolsByTrickyMap(data []byte) []byte {
	for _, r := range jsonSymbolReplacementsMap {
		data = bytes.ReplaceAll(data, r.from, r.to)
	}

	return data
}
