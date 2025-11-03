package api

import (
	"github.com/google/uuid"
)

type PatchImageRequest struct {
	ImageID uuid.UUID `json:"imageId"`
}
