package tools

import (
	"context"

	"github.com/distr-sh/distr/internal/types"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (m *Manager) NewListApplicationLicensesTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"list_application_licenses",
			mcp.WithDescription("This tool retrieves all application licenses"),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if licenses, err := m.client.ApplicationLicenses().List(ctx); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to list Application Licenses", err), nil
			} else {
				return JsonToolResult(licenses)
			}
		},
	}
}

func (m *Manager) NewGetApplicationLicenseTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"get_application_license",
			mcp.WithDescription("This tool retrieves a specific application license"),
			mcp.WithString("id", mcp.Required(), mcp.Description("ID of the license to retrieve")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := ParseUUID(request, "id")
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse license ID", err), nil
			}
			if id == uuid.Nil {
				return mcp.NewToolResultError("ID is required"), nil
			}

			if license, err := m.client.ApplicationLicenses().Get(ctx, id); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to get Application License", err), nil
			} else {
				return JsonToolResult(license)
			}
		},
	}
}

func (m *Manager) NewCreateApplicationLicenseTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"create_application_license",
			mcp.WithDescription("This tool creates a new application license"),
			mcp.WithObject("license", mcp.Required(), mcp.Description("License object to create")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			license, err := ParseT[*types.ApplicationLicenseWithVersions](request, "license", nil)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse license data", err), nil
			}

			if result, err := m.client.ApplicationLicenses().Create(ctx, license); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to create Application License", err), nil
			} else {
				return JsonToolResult(result)
			}
		},
	}
}

func (m *Manager) NewUpdateApplicationLicenseTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"update_application_license",
			mcp.WithDescription("This tool updates an existing application license"),
			mcp.WithObject("license", mcp.Required(), mcp.Description("License object to update")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			license, err := ParseT[*types.ApplicationLicenseWithVersions](request, "license", nil)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse license data", err), nil
			}

			if result, err := m.client.ApplicationLicenses().Update(ctx, license); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to update Application License", err), nil
			} else {
				return JsonToolResult(result)
			}
		},
	}
}

func (m *Manager) NewDeleteApplicationLicenseTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"delete_application_license",
			mcp.WithDescription("This tool deletes an application license"),
			mcp.WithString("id", mcp.Required(), mcp.Description("ID of the license to delete")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := ParseUUID(request, "id")
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse license ID", err), nil
			}
			if id == uuid.Nil {
				return mcp.NewToolResultError("ID is required"), nil
			}

			if err := m.client.ApplicationLicenses().Delete(ctx, id); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to delete Application License", err), nil
			}
			return JsonToolResult(map[string]string{"status": "Application License deleted successfully"})
		},
	}
}
