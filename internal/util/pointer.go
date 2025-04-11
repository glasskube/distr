package util

func PtrTo[T any](value T) *T {
	return &value
}

func PtrCopy[T any](ptr *T) *T {
	if ptr == nil {
		return nil
	}
	v := *ptr
	return &v
}
