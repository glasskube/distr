package tools

import (
	"context"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (m *Manager) NewCreateAccessForDeploymentTargetTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"create_access_for_deployment_target",
			mcp.WithDescription("This tool creates access credentials for a deployment target"),
			mcp.WithString("id", mcp.Required(), mcp.Description("ID of the deployment target")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := ParseUUID(request, "id")
			if err != nil {
				return mcp.NewToolResultErrorFromErr("id is invalid", err), nil
			}
			if id == uuid.Nil {
				return mcp.NewToolResultError("id is required"), nil
			}
			if accessResponse, err := m.client.DeploymentTargets().Connect(ctx, id); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to create access for DeploymentTarget", err), nil
			} else {
				return JsonToolResult(accessResponse)
			}
		},
	}
}
