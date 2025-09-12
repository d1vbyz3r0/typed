package typing

func EraseSliceType[T any](s []T) []any {
	res := make([]any, 0, len(s))
	for _, v := range s {
		res = append(res, v)
	}
	return res
}
