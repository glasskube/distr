package api

import (
	"github.com/google/uuid"
)

type PatchImageRequest struct {
	ImageID uuid.UUID `json:"imageId"`
}

func WithImageUrl(imageID *uuid.UUID) string {
	if imageID == nil || uuid.Nil == *imageID {
		return ""
	}
	return "/api/v1/files/" + imageID.String()
}
