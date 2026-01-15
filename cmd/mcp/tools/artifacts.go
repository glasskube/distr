package tools

import (
	"context"

	"github.com/distr-sh/distr/cmd/mcp/client"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (m *Manager) NewListArtifactsTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"list_artifacts",
			mcp.WithDescription("This tool retrieves all artifacts"),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if artifacts, err := m.client.Artifacts().List(ctx); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to list Artifacts", err), nil
			} else {
				return JsonToolResult(artifacts)
			}
		},
	}
}

func (m *Manager) NewGetArtifactTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"get_artifact",
			mcp.WithDescription("This tool retrieves a specific artifact"),
			mcp.WithString("id", mcp.Required(), mcp.Description("ID of the artifact to retrieve")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := ParseUUID(request, "id")
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse artifact ID", err), nil
			}
			if id == uuid.Nil {
				return mcp.NewToolResultError("ID is required"), nil
			}

			if artifact, err := m.client.Artifacts().Get(ctx, id); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to get Artifact", err), nil
			} else {
				return JsonToolResult(artifact)
			}
		},
	}
}

func (m *Manager) NewUpdateArtifactImageTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"update_artifact_image",
			mcp.WithDescription("This tool updates an artifact's image"),
			mcp.WithString("artifactId", mcp.Required(), mcp.Description("ID of the artifact")),
			mcp.WithString("imageId", mcp.Required(), mcp.Description("ID of the image")),
		),
		Handler: m.patchImageHandlerFunc(
			"artifactId",
			"imageId",
			func(c *client.Client) func(context.Context, uuid.UUID, uuid.UUID) (any, error) {
				return func(ctx context.Context, u1, u2 uuid.UUID) (any, error) {
					if result, err := c.Artifacts().UpdateImage(ctx, u1, u2); err != nil {
						return nil, err
					} else {
						return result, nil
					}
				}
			},
			"Failed to update Artifact image",
		),
	}
}
