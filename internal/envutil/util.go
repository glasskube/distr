package envutil

import (
	"fmt"
	"os"
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

func GetEnvParsedOrNil[T any](key string, parseFunc func(string) (T, error)) *T {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := parseFunc(value); err != nil {
			panic(fmt.Sprintf("malformed environment variable %v: %v", key, err))
		} else {
			return &parsed
		}
	}
	return nil
}

func GetEnvParsedOrDefault[T any](key string, parseFunc func(string) (T, error), defaultValue T) T {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := parseFunc(value); err != nil {
			panic(fmt.Sprintf("malformed environment variable %v: %v", key, err))
		} else {
			return parsed
		}
	}
	return defaultValue
}

func RequireEnv(key string) string {
	if value := GetEnv(key); value != "" {
		return value
	}
	panic(fmt.Sprintf("missing required environment variable: %v", key))
}

func RequireEnvParsed[T any](key string, parseFunc func(string) (T, error)) T {
	if parsed, err := parseFunc(RequireEnv(key)); err != nil {
		panic(fmt.Sprintf("malformed environment variable %v: %v", key, err))
	} else {
		return parsed
	}
}
