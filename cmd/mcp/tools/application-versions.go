package tools

import (
	"context"

	"github.com/distr-sh/distr/cmd/mcp/client"
	"github.com/distr-sh/distr/internal/types"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (m *Manager) NewGetApplicationVersionTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"get_application_version",
			mcp.WithDescription("This tool retrieves a specific application version"),
			mcp.WithString("applicationId", mcp.Required(), mcp.Description("ID of the application")),
			mcp.WithString("versionId", mcp.Required(), mcp.Description("ID of the version to retrieve")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			appID, err := ParseUUID(request, "applicationId")
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse application ID", err), nil
			}
			if appID == uuid.Nil {
				return mcp.NewToolResultError("Application ID is required"), nil
			}

			versionID, err := ParseUUID(request, "versionId")
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse version ID", err), nil
			}
			if versionID == uuid.Nil {
				return mcp.NewToolResultError("Version ID is required"), nil
			}

			if version, err := m.client.ApplicationVersions(appID).Get(ctx, versionID); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to get Application Version", err), nil
			} else {
				return JsonToolResult(version)
			}
		},
	}
}

func (m *Manager) NewCreateApplicationVersionTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"create_application_version",
			mcp.WithDescription("This tool creates a new application version"),
			mcp.WithString("applicationId", mcp.Required(), mcp.Description("ID of the application")),
			mcp.WithObject("version", mcp.Required(), mcp.Description("Version object to create")),
		),
		Handler: m.applicationVersionCreateUpdateFunc(
			func(av *client.ApplicationVersions) func(
				context.Context, types.ApplicationVersion) (*types.ApplicationVersion, error) {
				return av.Create
			},
			"Failed to create Application Version",
		),
	}
}

func (m *Manager) NewUpdateApplicationVersionTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"update_application_version",
			mcp.WithDescription("This tool updates an existing application version"),
			mcp.WithString("applicationId", mcp.Required(), mcp.Description("ID of the application")),
			mcp.WithObject("version", mcp.Required(), mcp.Description("Version object to update")),
		),
		Handler: m.applicationVersionCreateUpdateFunc(
			func(av *client.ApplicationVersions) func(
				context.Context, types.ApplicationVersion) (*types.ApplicationVersion, error) {
				return av.Update
			},
			"Failed to update Application Version",
		),
	}
}

func (m *Manager) applicationVersionCreateUpdateFunc(
	op func(*client.ApplicationVersions) func(
		context.Context, types.ApplicationVersion) (*types.ApplicationVersion, error),
	errorMessage string,
) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		appID, err := ParseUUID(request, "applicationId")
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to parse application ID", err), nil
		}

		version, err := ParseT[*types.ApplicationVersion](request, "version", nil)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to parse version data", err), nil
		}

		if result, err := op(m.client.ApplicationVersions(appID))(ctx, *version); err != nil {
			return mcp.NewToolResultErrorFromErr(errorMessage, err), nil
		} else {
			return JsonToolResult(result)
		}
	}
}
