package mapping

import (
	"github.com/distr-sh/distr/internal/util"
	"github.com/google/uuid"
)

func CreateImageURL(imageID *uuid.UUID) *string {
	if imageID == nil || *imageID == uuid.Nil {
		return nil
	}
	return util.PtrTo("/api/v1/files/" + imageID.String())
}
