package util

import (
	"errors"
	"fmt"
	"strings"
)

type mergeError struct {
	key     string
	wrapped error
}

func (err *mergeError) Error() string {
	path := []string{err.key}
	var root error
	next := err.wrapped
	for {
		if nm, ok := next.(*mergeError); ok {
			path = append(path, nm.key)
			next = nm.wrapped
		} else {
			root = next
			break
		}
	}
	return fmt.Sprintf("merge error at %v: %v", strings.Join(path, "."), root)
}

func MergeAllRecursive(mm ...map[string]any) (map[string]any, error) {
	dst := map[string]any{}
	for _, src := range mm {
		if err := MergeIntoRecursive(dst, src); err != nil {
			return nil, err
		}
	}
	return dst, nil
}

func MergeIntoRecursive(dst, src map[string]any) error {
	for k, sv := range src {
		if svm, ok := sv.(map[string]any); ok {
			if _, ok := dst[k].(map[string]any); !ok && dst[k] != nil {
				return &mergeError{key: k, wrapped: errors.New("can not merge map into non-map type")}
			}
			if dst[k] == nil {
				dst[k] = map[string]any{}
			}
			if err := MergeIntoRecursive(dst[k].(map[string]any), svm); err != nil {
				return &mergeError{key: k, wrapped: err}
			}
		} else if _, ok := dst[k].(map[string]any); ok {
			return &mergeError{key: k, wrapped: errors.New("can not merge non-map type into map")}
		} else {
			dst[k] = sv
		}
	}
	return nil
}
