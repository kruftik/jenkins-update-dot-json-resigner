package jenkins_update_center

import (
	"bytes"
	//"encoding/json"

	//"encoding/json"

	//"encoding/json"
	"fmt"
	cjson "github.com/gibson042/canonicaljson-go"
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

func extractJSONDocument(s []byte) ([]byte, error) {
	idxFrom := bytes.Index(s, []byte(`{`))
	idxTo := bytes.LastIndex(s, []byte(`}`))

	if idxFrom == -1 || idxTo == -1 {
		return nil, fmt.Errorf("cannot find a valid JSON document in the provided string")
	}

	return s[idxFrom : idxTo+1], nil
}

func prepareUpdateJSONObject(src []byte) (*UpdateJSON, error) {
	jsonStr, err := extractJSONDocument(src)
	if err != nil {
		return nil, fmt.Errorf("cannot strip json wrapping trailers: %s", err)
	}

	uj := &UpdateJSON{}

	err = cjson.Unmarshal(jsonStr, uj)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal update-center.json into struct: %s", err)
	}

	return uj, nil
}

func GetJSONPString(juc *UpdateJSON) ([]byte, error) {
	in, err := cjson.Marshal(juc)
	if err != nil {
		return nil, err
	}

	jsonp := make([]byte, 0, len(wrappedJSONPrefix)+len(in)+len(wrappedJSONPostfix))
	jsonp = append(jsonp, wrappedJSONPrefix...)
	jsonp = append(jsonp, in...)
	jsonp = append(jsonp, wrappedJSONPostfix...)

	return jsonp, nil
}

func getUnsignedJSON(signedObj UpdateJSON) ([]byte, error) {
	c := InsecureUpdateJSON(signedObj)

	data, err := cjson.Marshal(c)
	if err != nil {
		return nil, err
	}

	return replaceSymbolsByTrickyMap(data), nil
}

func (jsonData *InsecureUpdateJSON) GetBytes() ([]byte, error) {
	return cjson.Marshal(*jsonData)
}
