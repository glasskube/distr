package types

import (
	"fmt"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/google/uuid"
)

var (
	minVersionMultiDeployment = semver.MustParse("1.6.0")
)

type AgentVersion struct {
	ID                   uuid.UUID `db:"id" json:"id"`
	CreatedAt            time.Time `db:"created_at" json:"createdAt"`
	Name                 string    `db:"name" json:"name"`
	ManifestFileRevision string    `db:"manifest_file_revision" json:"-"`
	ComposeFileRevision  string    `db:"compose_file_revision" json:"-"`
}

func (av AgentVersion) CheckMultiDeploymentSupported() error {
	if av.Name == "snapshot" {
		return nil
	}
	sv, err := semver.NewVersion(av.Name)
	if err != nil {
		return err
	}
	if sv.LessThan(minVersionMultiDeployment) {
		return fmt.Errorf("multi deployments not supported by agent version %v (requires %v)", sv, minVersionMultiDeployment)
	}
	return nil
}
