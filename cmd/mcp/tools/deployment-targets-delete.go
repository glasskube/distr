package tools

import (
	"context"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (m *Manager) NewDeleteDeploymentTargetTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"delete_deployment_target",
			mcp.WithDescription("This tool deletes a deployment target"),
			mcp.WithString("id", mcp.Required(), mcp.Description("ID of the deployment target to delete")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := ParseUUID(request, "id")
			if err != nil {
				return mcp.NewToolResultErrorFromErr("id is invalid", err), nil
			}
			if id == uuid.Nil {
				return mcp.NewToolResultError("id is required"), nil
			}
			if err := m.client.DeploymentTargets().Delete(ctx, id); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to delete DeploymentTarget", err), nil
			} else {
				return JsonToolResult(map[string]string{"status": "success"})
			}
		},
	}
}
