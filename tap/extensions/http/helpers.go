package main

func rebuildHeaders(harHeaders interface{}) (headers map[string]interface{}) {
	headers = make(map[string]interface{})
	for _, header := range harHeaders.([]interface{}) {
		h := header.(map[string]interface{})
		headers[h["name"].(string)] = h["value"]
	}

	return
}
