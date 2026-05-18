package post

func ReverseSlice(s []string) []string {
	var result []string
	for i := len(s) - 1; i >= 0; i-- {
		result = append(result, s[i])
	}
	return result
}

func UniqueSlice(s []string) []string {
	var result []string
	var mark = make(map[string]struct{})
	for _, v := range s {
		if _, ok := mark[v]; !ok {
			mark[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}
