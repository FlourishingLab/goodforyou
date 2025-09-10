package shared

func MapToSlice(m map[int]Question) []Question {
	var slice []Question
	for _, v := range m {
		slice = append(slice, v)
	}
	return slice
}

func GetKeysFromMap(m map[string]Dimension) []string {
	keys := make([]string, 0, len(m)) // Preallocate slice with map size
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
