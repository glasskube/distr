package types

import "time"

type ArtifactVersionPull struct {
	CreatedAt            time.Time             `json:"createdAt"`
	RemoteAddress        *string               `json:"remoteAddress,omitempty"`
	UserAccount          *UserAccount          `json:"userAccount,omitempty"`
	CustomerOrganization *CustomerOrganization `json:"customerOrganization,omitempty"`
	Artifact             Artifact              `json:"artifact"`
	ArtifactVersion      ArtifactVersion       `json:"artifactVersion"`
}
