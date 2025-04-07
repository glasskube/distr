package types

import "time"

type TutorialProgressStep struct {
	Tutorial  Tutorial  `db:"tutorial" json:"tutorial"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type TutorialProgress struct {
	Tutorial  Tutorial  `db:"tutorial" json:"tutorial"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	Data      any       `db:"data" json:"data,omitempty"`
}
