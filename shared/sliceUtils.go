package shared

func Contains(slice []string, containsValue string) bool {
	for _, sliceValue := range slice {
		if sliceValue == containsValue {
			return true
		}
	}

	return false
}

func ContainsInt(slice []int, containsValue int) bool {
	for _, sliceValue := range slice {
		if sliceValue == containsValue {
			return true
		}
	}
	return false
}

func Unique(slice []string) []string {
	keys := make(map[string]bool)
	var list []string

	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	return list
}

func EqualStringSlices(slice1 []string, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for _, v := range slice1 {
		if !Contains(slice2, v) {
			return false
		}
	}

	return true
}
