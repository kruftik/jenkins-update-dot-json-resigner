package jenkins_update_center

//func downloadUpdateJSONFromURL(downloadURL string) (*json_schema.UpdateJSON, error) {
//	log.Infof("Downloading %s...", downloadURL)
//
//	resp, err := http.Get(downloadURL)
//	if err != nil {
//		return nil, fmt.Errorf("cannot GET %s: %s", downloadURL, err)
//	}
//	defer func() {
//		_ = resp.Body.Close()
//	}()
//
//	jsonFileData := &bytes.Buffer{}
//
//	n, err := jsonFileData.ReadFrom(resp.Body)
//	if err != nil {
//		return nil, fmt.Errorf("cannot save update.json content to buffer: %s", err)
//	}
//
//	log.Debugf("Successfully written %d bytes to buffer", n)
//
//	jsonStr, err := extractJSONDocument(jsonFileData.Bytes())
//	if err != nil {
//		return nil, fmt.Errorf("cannot strip json wrapping trailers: %s", err)
//	}
//
//	uj := &json_schema.UpdateJSON{}
//
//	err = json.Unmarshal(jsonStr, uj)
//	if err != nil {
//		return nil, fmt.Errorf("cannot unmarshal update-center.json into struct: %s", err)
//	}
//
//	return uj, nil
//}
//
//func readUpdateJSONFromFile(path string) (*json_schema.UpdateJSON, error) {

//}

//func NewUpdateJSONFromURL(downloadURL string) (*UpdateJSONT, error) {
//	f, err := downloadUpdateJSONFromURL(downloadURL)
//	if err != nil {
//		return nil, err
//	}
//	defer func() {
//		tempFileName := f.Name()
//		_ = f.Close()
//
//		if err := os.Remove(tempFileName); err != nil {
//			log.Errorf("Cannot delete obsolete temp file %s: %s", tempFileName, err)
//		}
//		log.Infof("Temp file %s successfully deleted", tempFileName)
//	}()
//
//	return NewUpdateJSONFromFile(f.Name())
//}
//
//func NewUpdateJSONFromFile(path string) (*UpdateJSONT, error) {
//	sbytes, err := ioutil.ReadFile(path)
//	if err != nil {
//		log.Error("Cannot read update.json content: ", err)
//		return nil, err
//	}
//
//
//
//	UpdateJSON := &json_schema.UpdateJSON{}
//	err = json.Unmarshal([]byte(jsonStr), UpdateJSON)
//	if err != nil {
//		log.Error("Cannot unmarshal update.json: ", err)
//		return nil, err
//	}
//
//	updateJSON := &UpdateJSONT{
//		json:      UpdateJSON,
//		Signature: &UpdateJSON.Signature,
//	}
//
//	return updateJSON, nil
//}
