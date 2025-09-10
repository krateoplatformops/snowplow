package api

func mapDepth(data any) int {
	switch v := data.(type) {
	case map[string]any:
		maxDepth := 1
		for _, val := range v {
			d := mapDepth(val) + 1
			if d > maxDepth {
				maxDepth = d
			}
		}
		return maxDepth
	case []any:
		maxDepth := 0
		for _, elem := range v {
			d := mapDepth(elem)
			if d > maxDepth {
				maxDepth = d
			}
		}
		return maxDepth
	default:
		return 0
	}
}
