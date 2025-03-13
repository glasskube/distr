package audit

import (
	"context"

	"github.com/glasskube/distr/internal/auth"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/registry/name"
)

type ArtifactAuditor interface {
	AuditPull(ctx context.Context, name, reference string) error
}

type auditor struct{}

func NewAuditor() ArtifactAuditor {
	return &auditor{}
}

// AuditPull implements ArtifactAuditor.
func (a *auditor) AuditPull(ctx context.Context, nameStr string, reference string) error {
	auth := auth.ArtifactsAuthentication.Require(ctx)
	if name, err := name.Parse(nameStr); err != nil {
		return err
	} else if version, err := db.GetArtifactVersion(ctx, name.OrgName, name.ArtifactName, reference); err != nil {
		return err
	} else if hasChilden, err := db.CheckArtifactVersionHasChildren(ctx, version.ID); err != nil {
		return err
	} else if !hasChilden {
		return db.CreateArtifactPullLogEntry(ctx, version.ID, auth.CurrentUserID())
	} else {
		return nil
	}
}
