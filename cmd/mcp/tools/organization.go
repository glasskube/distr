package tools

import (
	"context"

	"github.com/distr-sh/distr/internal/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (m *Manager) NewGetCurrentOrganizationTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"get_current_organization",
			mcp.WithDescription("This tool retrieves the current organization"),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if org, err := m.client.Organization().Current(ctx); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to get current Organization", err), nil
			} else {
				return JsonToolResult(org)
			}
		},
	}
}

func (m *Manager) NewCreateOrganizationTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"create_organization",
			mcp.WithDescription("This tool creates a new organization"),
			mcp.WithObject("organization", mcp.Required(), mcp.Description("Organization object to create")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			org, err := ParseT[*types.Organization](request, "organization", nil)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse organization data", err), nil
			}

			if result, err := m.client.Organization().Create(ctx, org); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to create Organization", err), nil
			} else {
				return JsonToolResult(result)
			}
		},
	}
}

func (m *Manager) NewUpdateOrganizationTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"update_organization",
			mcp.WithDescription("This tool updates an existing organization"),
			mcp.WithObject("organization", mcp.Required(), mcp.Description("Organization object to update")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			org, err := ParseT[*types.Organization](request, "organization", nil)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse organization data", err), nil
			} else if org == nil {
				return mcp.NewToolResultError("Organization data is required"), nil
			}

			if result, err := m.client.Organization().Update(ctx, org); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to update Organization", err), nil
			} else {
				return JsonToolResult(result)
			}
		},
	}
}
