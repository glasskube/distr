package util

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Require[T any](v T, err error) T {
	Must(err)
	return v
}
