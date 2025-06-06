package tools

import (
	"context"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (m *Manager) NewGetDeploymentTargetsTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"get_deployment_target",
			mcp.WithDescription("This tools retrieves a particular deployment target with the specified ID"),
			mcp.WithString("id", mcp.Required(), mcp.Description("ID of the deployment target")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if idStr := mcp.ParseString(request, "id", ""); idStr == "" {
				return mcp.NewToolResultError("id is required"), nil
			} else if id, err := uuid.Parse(idStr); err != nil {
				return mcp.NewToolResultErrorFromErr("id is invalid", err), nil
			} else if deploymentTargets, err := m.client.DeploymentTargets().Get(ctx, id); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to get DeploymentTarget", err), nil
			} else {
				return JsonToolResult(deploymentTargets)
			}
		},
	}
}
