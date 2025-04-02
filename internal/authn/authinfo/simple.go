package authinfo

import (
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
)

type SimpleAuthInfo struct {
	userID         uuid.UUID
	userEmail      string
	organizationID *uuid.UUID
	emailVerified  bool
	userRole       *types.UserRole
	rawToken       any
}

// CurrentOrgID implements AuthInfo.
func (i *SimpleAuthInfo) CurrentOrgID() *uuid.UUID { return i.organizationID }

// CurrentUserEmailVerified implements AuthInfo.
func (i *SimpleAuthInfo) CurrentUserEmailVerified() bool { return i.emailVerified }

// CurrentUserID implements AuthInfo.
func (i *SimpleAuthInfo) CurrentUserID() uuid.UUID { return i.userID }

// CurrentUserEmail implements AuthInfo.
func (i *SimpleAuthInfo) CurrentUserEmail() string { return i.userEmail }

// CurrentUserRole implements AuthInfo.
func (i *SimpleAuthInfo) CurrentUserRole() *types.UserRole { return i.userRole }

// Token implements AuthInfo.
func (i *SimpleAuthInfo) Token() any { return i.rawToken }

type SimpleAgentAuthInfo struct {
	deploymentTargetID uuid.UUID
	organizationID     uuid.UUID
	rawToken           any
}

// CurrentDeploymentTargetID implements AgentAuthInfo.
func (i *SimpleAgentAuthInfo) CurrentDeploymentTargetID() uuid.UUID {
	return i.deploymentTargetID
}

// CurrentOrgID implements AgentAuthInfo.
func (i *SimpleAgentAuthInfo) CurrentOrgID() uuid.UUID {
	return i.organizationID
}

// Token implements AgentAuthInfo.
func (i *SimpleAgentAuthInfo) Token() any {
	return i.rawToken
}
