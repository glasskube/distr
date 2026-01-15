package tools

import (
	"context"
	"fmt"

	"github.com/distr-sh/distr/cmd/mcp/client"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (m *Manager) patchImageHandlerFunc(
	objIDKey string,
	imageIDKey string,
	op func(*client.Client) func(context.Context, uuid.UUID, uuid.UUID) (any, error),
	errorMessage string,
) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		objID, err := ParseUUID(request, objIDKey)
		if err != nil {
			return mcp.NewToolResultErrorFromErr(fmt.Sprintf("Failed to parse %v", objIDKey), err), nil
		}

		imageID, err := ParseUUID(request, imageIDKey)
		if err != nil {
			return mcp.NewToolResultErrorFromErr(fmt.Sprintf("Failed to parse %v", imageIDKey), err), nil
		}

		if result, err := op(m.client)(ctx, objID, imageID); err != nil {
			return mcp.NewToolResultErrorFromErr(errorMessage, err), nil
		} else {
			return JsonToolResult(result)
		}
	}
}
