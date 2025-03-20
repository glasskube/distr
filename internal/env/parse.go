package env

import (
	"errors"
	"net/mail"
	"time"
)

func getPositiveDuration(value string) (time.Duration, error) {
	parsed, err := time.ParseDuration(value)
	if err == nil && parsed.Nanoseconds() <= 0 {
		err = errors.New("duration must be positive")
	}
	return parsed, err
}

func asByteSlice(s string) ([]byte, error) {
	return []byte(s), nil
}

func parseMailAddress(s string) (mail.Address, error) {
	if parsed, err := mail.ParseAddress(s); err != nil || parsed == nil {
		return mail.Address{}, err
	} else {
		return *parsed, nil
	}
}
