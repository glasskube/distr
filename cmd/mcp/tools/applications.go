package tools

import (
	"context"
	"encoding/json"

	"github.com/distr-sh/distr/internal/types"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (m *Manager) NewListApplicationsTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"list_applications",
			mcp.WithDescription("This tool retrieves a list of all applications"),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if applications, err := m.client.Applications().List(ctx); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to list Applications", err), nil
			} else {
				return JsonToolResult(applications)
			}
		},
	}
}

func (m *Manager) NewGetApplicationTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"get_application",
			mcp.WithDescription("This tool retrieves a specific application by ID"),
			mcp.WithString("id", mcp.Required(), mcp.Description("ID of the application to retrieve")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := ParseUUID(request, "id")
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse application ID", err), nil
			}
			if id == uuid.Nil {
				return mcp.NewToolResultError("Application ID is required"), nil
			}

			if application, err := m.client.Applications().Get(ctx, id); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to get Application", err), nil
			} else {
				return JsonToolResult(application)
			}
		},
	}
}

func (m *Manager) NewCreateApplicationTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"create_application",
			mcp.WithDescription("This tool creates a new application"),
			mcp.WithObject("application", mcp.Required(), mcp.Description("Application object to create")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			app, err := ParseT[*types.Application](request, "application", nil)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse application data", err), nil
			} else if app == nil {
				return mcp.NewToolResultErrorFromErr("application is required", err), nil
			}

			if result, err := m.client.Applications().Create(ctx, *app); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to create Application", err), nil
			} else {
				return JsonToolResult(result)
			}
		},
	}
}

func (m *Manager) NewUpdateApplicationTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"update_application",
			mcp.WithDescription("This tool updates an existing application"),
			mcp.WithObject("application", mcp.Required(), mcp.Description("Application object to update")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var defaultAppData map[string]any
			appData := mcp.ParseStringMap(request, "application", defaultAppData)
			if appData == nil {
				return mcp.NewToolResultError("Application data is required"), nil
			}

			appJSON, err := json.Marshal(appData)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to process application data", err), nil
			}

			var app types.Application
			if err := json.Unmarshal(appJSON, &app); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse application data", err), nil
			}

			if result, err := m.client.Applications().Update(ctx, app); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to update Application", err), nil
			} else {
				return JsonToolResult(result)
			}
		},
	}
}
