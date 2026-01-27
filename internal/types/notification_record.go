package types

import (
	"time"

	"github.com/google/uuid"
)

type NotificationRecord struct {
	ID                                          uuid.UUID  `db:"id" json:"id"`
	CreatedAt                                   time.Time  `db:"created_at" json:"createdAt"`
	DeploymentStatusNotificationConfigurationID *uuid.UUID `db:"deployment_status_notification_configuration_id" json:"deploymentStatusNotificationConfigurationId"` //nolint:lll
	PreviousDeploymentStatusID                  *uuid.UUID `db:"previous_deployment_revision_status_id" json:"previousDeploymentStatusId"`                           //nolint:lll
	CurrentDeploymentStatusID                   *uuid.UUID `db:"current_deployment_status_id" json:"currentDeploymentStatusId"`                                      //nolint:lll
	Message                                     string     `db:"message" json:"message"`
}
