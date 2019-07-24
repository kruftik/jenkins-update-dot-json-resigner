package jenkins_update_center

import (
	"bytes"
	"encoding/json"
	"fmt"
	"jenkins-resigner-service/jenkins_update_center/json_schema"
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

func prepareUpdateJSONObject(src []byte) (*json_schema.UpdateJSON, error) {
	jsonStr, err := extractJSONDocument(src)
	if err != nil {
		return nil, fmt.Errorf("cannot strip json wrapping trailers: %s", err)
	}

	uj := &json_schema.UpdateJSON{}

	err = json.Unmarshal(jsonStr, uj)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal update-center.json into struct: %s", err)
	}

	return uj, nil
}

func getJSONPString(jsp JSONProvider) ([]byte, error) {
	content, err := jsp.GetContent()
	if err != nil {
		return nil, err
	}

	in, err := json.Marshal(content)

	jsonp := make([]byte, 0, len(wrappedJSONPrefix)+len(in)+len(wrappedJSONPostfix))
	jsonp = append(jsonp, wrappedJSONPrefix...)
	jsonp = append(jsonp, in...)
	jsonp = append(jsonp, wrappedJSONPostfix...)

	return jsonp, nil
}

func getUnsignedJSON(signedObj json_schema.UpdateJSON) ([]byte, error) {
	c := json_schema.InsecureUpdateJSON(signedObj)

	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	return replaceSymbolsByTrickyMap(data), nil
}
