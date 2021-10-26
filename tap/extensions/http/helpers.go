package main

func rebuildAsMap(harMap interface{}) (newMap map[string]interface{}) {
	newMap = make(map[string]interface{})
	for _, header := range harMap.([]interface{}) {
		h := header.(map[string]interface{})
		newMap[h["name"].(string)] = h["value"]
	}

	return
}
