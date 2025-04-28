package apierrors

import "errors"

var ErrNotFound = errors.New("not found")
var ErrAlreadyExists = errors.New("already exists")
var ErrConflict = errors.New("conflict")
var ErrForbidden = errors.New("forbidden")
var ErrQuotaExceeded = errors.New("quota exceeded")
