package api

import "github.com/glasskube/distr/internal/types"

type TutorialProgressRequest struct {
	types.TutorialProgressEvent
	MarkCompleted bool `json:"markCompleted"`
}
