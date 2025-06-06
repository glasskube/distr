package tools

import (
	"context"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (m *Manager) NewGetFileTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"get_file",
			mcp.WithDescription("This tool retrieves a file by ID"),
			mcp.WithString("id", mcp.Required(), mcp.Description("ID of the file to retrieve")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			idStr := mcp.ParseString(request, "id", "")
			if idStr == "" {
				return mcp.NewToolResultError("ID is required"), nil
			}

			id, err := uuid.Parse(idStr)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse file ID", err), nil
			}

			if file, err := m.client.Files().Get(ctx, id); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to get File", err), nil
			} else {
				return JsonToolResult(file)
			}
		},
	}
}

func (m *Manager) NewDeleteFileTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"delete_file",
			mcp.WithDescription("This tool deletes a file by ID"),
			mcp.WithString("id", mcp.Required(), mcp.Description("ID of the file to delete")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			idStr := mcp.ParseString(request, "id", "")
			if idStr == "" {
				return mcp.NewToolResultError("ID is required"), nil
			}

			id, err := uuid.Parse(idStr)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse file ID", err), nil
			}

			if err := m.client.Files().Delete(ctx, id); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to delete File", err), nil
			} else {
				return JsonToolResult(map[string]string{"status": "success"})
			}
		},
	}
}
