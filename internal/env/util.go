package env

import (
	"fmt"
	"os"
)

type getEnvOpts struct {
	deprecatedAlias string
}

func getEnv(key string) string {
	return os.Getenv(key)
}

func getEnvOrNil(key string) *string {
	if value, ok := os.LookupEnv(key); ok {
		return &value
	}
	return nil
}

func getEnvOrDefault(key, defaultValue string, opts getEnvOpts) string {
	if value := getEnv(key); value != "" {
		return value
	} else if opts.deprecatedAlias != "" {
		if value := getEnv(opts.deprecatedAlias); value != "" {
			fmt.Fprintf(os.Stderr, "\nWARNING: use of deprecated variable \"%v\", please use \"%v\" instead\n\n",
				opts.deprecatedAlias, key)
			return value
		}
	}
	return defaultValue
}

func getEnvParsedOrNil[T any](key string, parseFunc func(string) (T, error)) *T {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := parseFunc(value); err != nil {
			panic(fmt.Sprintf("malformed environment variable %v: %v", key, err))
		} else {
			return &parsed
		}
	}
	return nil
}

func getEnvParsedOrDefault[T any](key string, parseFunc func(string) (T, error), defaultValue T) T {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := parseFunc(value); err != nil {
			panic(fmt.Sprintf("malformed environment variable %v: %v", key, err))
		} else {
			return parsed
		}
	}
	return defaultValue
}

func requireEnv(key string) string {
	if value := getEnv(key); value != "" {
		return value
	}
	panic(fmt.Sprintf("missing required environment variable: %v", key))
}

func requireEnvParsed[T any](key string, parseFunc func(string) (T, error)) T {
	if parsed, err := parseFunc(requireEnv(key)); err != nil {
		panic(fmt.Sprintf("malformed environment variable %v: %v", key, err))
	} else {
		return parsed
	}
}
