package authn

import "errors"

// ErrNoAuthentication implies that the provider *did not* find relevant
// authentication information on the Request
var ErrNoAuthentication = errors.New("not authenticated")

// ErrBadAuthentication implies that the provider *did* find relevant
// authentication information on the Request but it is not valid
var ErrBadAuthentication = errors.New("bad authentication")
