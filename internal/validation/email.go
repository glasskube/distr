package validation

import (
	"regexp"
)

var emailFormatPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) error {
	if !emailFormatPattern.MatchString(email) {
		return NewValidationFailedError("invalid email format")
	}
	return nil
}
