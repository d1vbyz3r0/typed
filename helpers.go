package typed

// MakePointer returns a pointer to v.
func MakePointer[T any](v T) *T {
	return &v
}
