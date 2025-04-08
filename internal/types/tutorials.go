package types

import "time"

type TutorialProgressEvent struct {
	StepID    string    `json:"stepId"`
	TaskID    string    `json:"taskId"`
	Value     any       `json:"value,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type TutorialProgressRequest struct {
	TutorialProgressEvent
	MarkCompleted bool `json:"markCompleted"`
}

type TutorialProgress struct {
	Tutorial    Tutorial                `db:"tutorial" json:"tutorial"`
	CreatedAt   time.Time               `db:"created_at" json:"createdAt"`
	Events      []TutorialProgressEvent `db:"events" json:"events,omitempty"`
	CompletedAt *time.Time              `db:"completed_at" json:"completedAt,omitempty"`
}
