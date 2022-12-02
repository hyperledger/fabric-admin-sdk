package util

func Concatenate[T any](slices ...[]T) []T {
	size := 0
	for _, slice := range slices {
		size += len(slice)
	}

	result := make([]T, size)
	i := 0
	for _, slice := range slices {
		copy(result[i:], slice)
		i += len(slice)
	}

	return result
}
