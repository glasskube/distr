package name

import (
	"fmt"
	"path"
	"strings"

	registryerror "github.com/distr-sh/distr/internal/registry/error"
)

type Name struct {
	OrgName      string
	ArtifactName string
}

func Parse(input string) (*Name, error) {
	if parts := strings.SplitN(input, "/", 2); len(parts) != 2 {
		return nil, fmt.Errorf("%w: %v", registryerror.ErrInvalidArtifactName, input)
	} else {
		return &Name{parts[0], parts[1]}, nil
	}
}

func (obj Name) String() string {
	return path.Join(obj.OrgName, obj.ArtifactName)
}
