package mizu

func Contains(slice []string, containsValue string) bool {
	for _, sliceValue := range slice {
		if sliceValue == containsValue {
			return true
		}
	}

	return false
}
