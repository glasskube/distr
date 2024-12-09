package util

func PtrTo[T any](value T) *T {
	return &value
}
