package validation

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return NewValidationFailedError("password is too short (minimum 8 characters are required)")
	}
	return nil
}
