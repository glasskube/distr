package tools

import (
	"github.com/glasskube/distr/cmd/mcp/client"
	"github.com/mark3labs/mcp-go/server"
)

type Manager struct {
	client *client.Client
}

func NewManager(client *client.Client) *Manager {
	return &Manager{client: client}
}

func (m *Manager) AddToolsToServer(mcpServer *server.MCPServer) {
	mcpServer.AddTools(
		// Deployment Target tools
		m.NewListDeploymentTargetsTool(),
		m.NewGetDeploymentTargetsTool(),
		m.NewCreateDeploymentTargetTool(),
		m.NewUpdateDeploymentTargetTool(),
		m.NewDeleteDeploymentTargetTool(),
		m.NewCreateAccessForDeploymentTargetTool(),

		// Application tools
		m.NewListApplicationsTool(),
		m.NewGetApplicationTool(),
		m.NewCreateApplicationTool(),
		m.NewUpdateApplicationTool(),

		// Application Version tools
		m.NewGetApplicationVersionTool(),
		m.NewCreateApplicationVersionTool(),
		m.NewUpdateApplicationVersionTool(),
		m.NewComposeFileTool(),
		m.NewValuesFileTool(),
		m.NewTemplateFileTool(),

		// Application License tools
		m.NewListApplicationLicensesTool(),
		m.NewGetApplicationLicenseTool(),
		m.NewCreateApplicationLicenseTool(),
		m.NewUpdateApplicationLicenseTool(),
		m.NewDeleteApplicationLicenseTool(),

		// User tools
		m.NewListUsersTool(),
		m.NewCreateUserTool(),
		m.NewDeleteUserTool(),
		m.NewUpdateUserImageTool(),

		// File tools
		m.NewGetFileTool(),
		m.NewDeleteFileTool(),

		// Artifact tools
		m.NewListArtifactsTool(),
		m.NewGetArtifactTool(),
		m.NewUpdateArtifactImageTool(),

		// Organization tools
		m.NewGetCurrentOrganizationTool(),
		m.NewCreateOrganizationTool(),
		m.NewUpdateOrganizationTool(),

		// Deployment tools
		m.NewCreateOrUpdateDeploymentTool(),
		m.NewPatchDeploymentTool(),
		m.NewDeleteDeploymentTool(),
		m.NewStatusTool(),
		m.NewLogsTool(),
		m.NewLogResourcesTool(),
	)
}
