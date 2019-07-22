package jenkins_update_center

import "bytes"

func replaceSymbolsByTrickyMap(data []byte) []byte {
	for _, r := range jsonSymbolReplacementsMap {
		data = bytes.ReplaceAll(data, r.from, r.to)
	}

	return data
}

