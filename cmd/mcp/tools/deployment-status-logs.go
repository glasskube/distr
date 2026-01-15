package tools

import (
	"context"
	"fmt"
	"io"

	"github.com/distr-sh/distr/cmd/mcp/client"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ComposeFile tool
func (m *Manager) NewComposeFileTool() server.ServerTool {
	return m.newFilesTool("get_compose_file", "docker compose",
		func(c *client.ApplicationVersions) func(context.Context, uuid.UUID) (io.ReadCloser, error) {
			return c.ComposeFile
		},
	)
}

// ValuesFile tool
func (m *Manager) NewValuesFileTool() server.ServerTool {
	return m.newFilesTool(
		"get_values_file",
		"values",
		func(c *client.ApplicationVersions) func(context.Context, uuid.UUID) (io.ReadCloser, error) {
			return c.ValuesFile
		},
	)
}

// TemplateFile tool
func (m *Manager) NewTemplateFileTool() server.ServerTool {
	return m.newFilesTool(
		"get_template_file",
		"template",
		func(c *client.ApplicationVersions) func(context.Context, uuid.UUID) (io.ReadCloser, error) {
			return c.TemplateFile
		},
	)
}

func (m *Manager) newFilesTool(
	toolName, fileName string,
	apiCallFunc func(*client.ApplicationVersions) func(context.Context, uuid.UUID) (io.ReadCloser, error),
) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			toolName,
			mcp.WithDescription(fmt.Sprintf("Get the %v file for an application version", fileName)),
			mcp.WithString("applicationId", mcp.Required(), mcp.Description("Application ID")),
			mcp.WithString("id", mcp.Required(), mcp.Description("Application version ID")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			applicationID, err := ParseUUID(request, "applicationId")
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse application ID", err), nil
			}
			id, err := ParseUUID(request, "id")
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse application version ID", err), nil
			}
			file, err := apiCallFunc(m.client.ApplicationVersions(applicationID))(ctx, id)
			if err != nil {
				return mcp.NewToolResultErrorFromErr(fmt.Sprintf("Failed to get %v file", fileName), err), nil
			}
			defer file.Close()
			data, err := io.ReadAll(file)
			if err != nil {
				return mcp.NewToolResultErrorFromErr(fmt.Sprintf("Failed to read %v file", fileName), err), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		},
	}
}
