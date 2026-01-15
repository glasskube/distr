package customdomains

import (
	"net/mail"
	"regexp"

	"github.com/distr-sh/distr/internal/env"
	"github.com/distr-sh/distr/internal/types"
	"github.com/distr-sh/distr/internal/util"
)

var urlSchemeRegex = regexp.MustCompile("^https?://")

func AppDomainOrDefault(o types.Organization) string {
	if o.AppDomain != nil {
		d := *o.AppDomain
		if urlSchemeRegex.MatchString(d) {
			return d
		} else {
			scheme := urlSchemeRegex.FindString(env.Host())
			if scheme == "" {
				scheme = "https://"
			}
			return scheme + d
		}
	} else {
		return env.Host()
	}
}

func RegistryDomainOrDefault(o types.Organization) string {
	if o.RegistryDomain != nil {
		return *o.RegistryDomain
	} else {
		return env.RegistryHost()
	}
}

func EmailFromAddressParsedOrDefault(o types.Organization) (*mail.Address, error) {
	if o.EmailFromAddress != nil {
		return mail.ParseAddress(*o.EmailFromAddress)
	} else {
		return util.PtrTo(env.GetMailerConfig().FromAddress), nil
	}
}
