package instructor

func toPtr[T any](val T) *T {
	return &val
}

func prepend[T any](to []T, from T) []T {
	return append([]T{from}, to...)
}
