package allmodels

type Page[T any] struct {
	Items []T
}

type KV[K comparable, V any] struct {
	Key   K
	Value V
}
