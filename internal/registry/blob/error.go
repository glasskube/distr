package blob

import (
	"errors"
	"fmt"
)

// RedirectError represents a signal that the blob handler doesn't have the blob
// contents, but that those contents are at another location which registry
// clients should redirect to.
type RedirectError struct {
	// Location is the location to find the contents.
	Location string

	// Code is the HTTP redirect status code to return to clients.
	Code int
}

func (e RedirectError) Error() string { return fmt.Sprintf("redirecting (%d): %s", e.Code, e.Location) }

// errNotFound represents an error locating the blob.
var ErrNotFound = errors.New("not found")

var ErrBadUpload = errors.New("bad upload")

func NewErrBadUpload(msg string) error {
	return fmt.Errorf("%w: %v", ErrBadUpload, msg)
}
