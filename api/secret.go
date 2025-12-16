package api

import (
	"time"

	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
)

type SecretWithoutValue struct {
	ID                     uuid.UUID          `json:"id"`
	CreatedAt              time.Time          `json:"createdAt"`
	UpdatedAt              time.Time          `json:"updatedAt"`
	UpdatedBy              *types.UserAccount `json:"updatedBy,omitempty"`
	CustomerOrganizationID *uuid.UUID         `json:"customerOrganizationId,omitempty"`
	Key                    string             `json:"key"`
}

type CreateUpdateSecretRequest struct {
	Key   string `json:"key" path:"key"`
	Value string `json:"value"`
}

type DeleteSecretRequest struct {
	Key string `json:"key" path:"key"`
}
