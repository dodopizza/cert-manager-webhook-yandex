package internal

func ContainsString(item string, items []string) bool {
	for _, s := range items {
		if s == item {
			return true
		}
	}
	return false
}
