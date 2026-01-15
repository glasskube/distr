package api

import "github.com/distr-sh/distr/internal/types"

type TutorialProgressRequest struct {
	types.TutorialProgressEvent
	MarkCompleted bool `json:"markCompleted"`
}
