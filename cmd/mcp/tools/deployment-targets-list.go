package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (m *Manager) NewListDeploymentTargetsTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"list_deployment_targets",
			mcp.WithDescription("This tools retrieves a list of all available DeploymentTargets"),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if deploymentTargets, err := m.client.DeploymentTargets().List(ctx); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to list DeploymentTargets", err), nil
			} else {
				return JsonToolResult(deploymentTargets)
			}
		},
	}
}
