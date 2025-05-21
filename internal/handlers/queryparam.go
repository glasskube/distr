package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

var ErrParamNotDefined = errors.New("parameter not defined")

func QueryParam[T any](r *http.Request, name string, parseFunc func(string) (T, error), validatorFunc ...func(T) error) (T, error) {
	if stringValue := r.FormValue(name); stringValue == "" {
		var v T
		return v, ErrParamNotDefined
	} else if value, err := parseFunc(stringValue); err != nil {
		return value, fmt.Errorf("parameter %v is invalid: %w", name, err)
	} else {
		for _, fn := range validatorFunc {
			if err := fn(value); err != nil {
				return value, fmt.Errorf("parameter %v is invalid: %w", name, err)
			}
		}
		return value, nil
	}
}

func ParseTimeFunc(layout string) func(string) (time.Time, error) {
	return func(value string) (time.Time, error) {
		return time.Parse(layout, value)
	}
}

func Min(min int) func(int) error {
	return func(v int) error {
		if v < min {
			return fmt.Errorf("must be greater than %v (got %v)", min, v)
		}
		return nil
	}
}

func Max(max int) func(int) error {
	return func(v int) error {
		if v > max {
			return fmt.Errorf("must be less than %v (got %v)", max, v)
		}
		return nil
	}
}
