package mapping

import (
	"github.com/distr-sh/distr/api"
	"github.com/distr-sh/distr/internal/types"
)

func DeploymentTargetNotesToAPI(notes types.DeploymentTargetNotes) api.DeploymentTargetNotes {
	return api.DeploymentTargetNotes{
		Notes:                  notes.Notes,
		UpdatedByUserAccountID: notes.UpdatedByUserAccountID,
		UpdatedAt:              &notes.UpdatedAt,
	}
}
