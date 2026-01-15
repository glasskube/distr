package envutil

import (
	"fmt"
	"os"

	"github.com/distr-sh/distr/internal/util"
)

type GetEnvOpts struct {
	DeprecatedAlias string
}

func GetEnv(key string) string {
	return os.Getenv(key)
}

func GetEnvOrNil(key string) *string {
	if value, ok := os.LookupEnv(key); ok {
		return &value
	}
	return nil
}

func GetEnvOrDefault(key, defaultValue string, opts GetEnvOpts) string {
	if value := GetEnv(key); value != "" {
		return value
	} else if opts.DeprecatedAlias != "" {
		if value := GetEnv(opts.DeprecatedAlias); value != "" {
			fmt.Fprintf(os.Stderr, "\nWARNING: use of deprecated variable \"%v\", please use \"%v\" instead\n\n",
				opts.DeprecatedAlias, key)
			return value
		}
	}
	return defaultValue
}

func GetEnvParsedOrNilErr[T any](key string, parseFunc func(string) (T, error)) (*T, error) {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := parseFunc(value); err != nil {
			return nil, fmt.Errorf("malformed environment variable %v: %v", key, err)
		} else {
			return &parsed, nil
		}
	}
	return nil, nil
}

func GetEnvParsedOrNil[T any](key string, parseFunc func(string) (T, error)) *T {
	return util.Require(GetEnvParsedOrNilErr(key, parseFunc))
}

func GetEnvParsedOrDefaultErr[T any](key string, parseFunc func(string) (T, error), defaultValue T) (T, error) {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := parseFunc(value); err != nil {
			return parsed, fmt.Errorf("malformed environment variable %v: %v", key, err)
		} else {
			return parsed, nil
		}
	}
	return defaultValue, nil
}

func GetEnvParsedOrDefault[T any](key string, parseFunc func(string) (T, error), defaultValue T) T {
	return util.Require(GetEnvParsedOrDefaultErr(key, parseFunc, defaultValue))
}

func RequireEnvErr(key string) (string, error) {
	if value := GetEnv(key); value != "" {
		return value, nil
	}
	return "", fmt.Errorf("missing required environment variable: %v", key)
}

func RequireEnv(key string) string {
	return util.Require(RequireEnvErr(key))
}

func RequireEnvParsedErr[T any](key string, parseFunc func(string) (T, error)) (T, error) {
	if value, err := RequireEnvErr(key); err != nil {
		var empty T
		return empty, err
	} else if parsed, err := parseFunc(value); err != nil {
		return parsed, fmt.Errorf("malformed environment variable %v: %v", key, err)
	} else {
		return parsed, nil
	}
}

func RequireEnvParsed[T any](key string, parseFunc func(string) (T, error)) T {
	return util.Require(RequireEnvParsedErr(key, parseFunc))
}
