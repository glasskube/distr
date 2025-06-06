package tools

import (
	"context"
	"encoding/json"

	"github.com/glasskube/distr/internal/types"
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
			appIDStr := mcp.ParseString(request, "applicationId", "")
			if appIDStr == "" {
				return mcp.NewToolResultError("Application ID is required"), nil
			}

			versionIDStr := mcp.ParseString(request, "versionId", "")
			if versionIDStr == "" {
				return mcp.NewToolResultError("Version ID is required"), nil
			}

			appID, err := uuid.Parse(appIDStr)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse application ID", err), nil
			}

			versionID, err := uuid.Parse(versionIDStr)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse version ID", err), nil
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
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			appIDStr := mcp.ParseString(request, "applicationId", "")
			if appIDStr == "" {
				return mcp.NewToolResultError("Application ID is required"), nil
			}

			var defaultVersionData map[string]any
			versionData := mcp.ParseStringMap(request, "version", defaultVersionData)
			if versionData == nil {
				return mcp.NewToolResultError("Version data is required"), nil
			}

			appID, err := uuid.Parse(appIDStr)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse application ID", err), nil
			}

			versionJSON, err := json.Marshal(versionData)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to process version data", err), nil
			}

			var version types.ApplicationVersion
			if err := json.Unmarshal(versionJSON, &version); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse version data", err), nil
			}

			if result, err := m.client.ApplicationVersions(appID).Create(ctx, version); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to create Application Version", err), nil
			} else {
				return JsonToolResult(result)
			}
		},
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
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			appIDStr := mcp.ParseString(request, "applicationId", "")
			if appIDStr == "" {
				return mcp.NewToolResultError("Application ID is required"), nil
			}

			var defaultVersionData map[string]any
			versionData := mcp.ParseStringMap(request, "version", defaultVersionData)
			if versionData == nil {
				return mcp.NewToolResultError("Version data is required"), nil
			}

			appID, err := uuid.Parse(appIDStr)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse application ID", err), nil
			}

			versionJSON, err := json.Marshal(versionData)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to process version data", err), nil
			}

			var version types.ApplicationVersion
			if err := json.Unmarshal(versionJSON, &version); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse version data", err), nil
			}

			if result, err := m.client.ApplicationVersions(appID).Update(ctx, version); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to update Application Version", err), nil
			} else {
				return JsonToolResult(result)
			}
		},
	}
}
