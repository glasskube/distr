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
		m.NewListDeploymentTargetsTool(),
		m.NewGetDeploymentTargetsTool(),
		m.NewCreateDeploymentTargetTool(),
	)
}
