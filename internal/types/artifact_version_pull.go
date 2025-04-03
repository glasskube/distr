package types

import "time"

type ArtifactVersionPull struct {
	CreatedAt       time.Time       `json:"createdAt"`
	RemoteAddress   *string         `json:"remoteAddress"`
	UserAccount     *UserAccount    `json:"userAccount"`
	Artifact        Artifact        `json:"artifact"`
	ArtifactVersion ArtifactVersion `json:"artifactVersion"`
}
