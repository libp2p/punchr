package util

func UniqueStr(input []string) *string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	if len(u) == 1 {
		return &u[0]
	}

	return nil
}

func UniqueInt(input []int) *int {
	u := make([]int, 0, len(input))
	m := make(map[int]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	if len(u) == 1 {
		return &u[0]
	}

	return nil
}
