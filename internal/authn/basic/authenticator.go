package basic

import (
	"context"
	"net/http"

	"github.com/glasskube/distr/internal/authn"
)

type Auth struct {
	Username, Password string
}

func Authenticator() authn.RequestAuthenticator[Auth] {
	fn := func(ctx context.Context, r *http.Request) (result Auth, err error) {
		if username, password, ok := r.BasicAuth(); !ok {
			err = authn.NewHttpHeaderError(
				authn.ErrNoAuthentication,
				http.Header{
					"WWW-Authenticate": []string{`Bearer realm="http://localhost:8585/v2/",service="localhost:8585"`, `Basic realm="Distr"`},
				},
			)
		} else {
			result = Auth{Username: username, Password: password}
		}
		return
	}
	return authn.AuthenticatorFunc[*http.Request, Auth](fn)
}

func Password() authn.Authenticator[Auth, string] {
	return authn.AuthenticatorFunc[Auth, string](func(ctx context.Context, basic Auth) (string, error) {
		return basic.Password, nil
	})
}
