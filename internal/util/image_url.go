package util

import "github.com/google/uuid"

func CreateImageURL(imageID *uuid.UUID) string {
	if imageID == nil || *imageID == uuid.Nil {
		return ""
	}
	return "/api/v1/files/" + imageID.String()
}
