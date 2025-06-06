package tools

import (
	"context"

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
			idStr := mcp.ParseString(request, "id", "")
			if idStr == "" {
				return mcp.NewToolResultError("ID is required"), nil
			}

			id, err := uuid.Parse(idStr)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse artifact ID", err), nil
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
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			artifactIDStr := mcp.ParseString(request, "artifactId", "")
			if artifactIDStr == "" {
				return mcp.NewToolResultError("Artifact ID is required"), nil
			}

			imageIDStr := mcp.ParseString(request, "imageId", "")
			if imageIDStr == "" {
				return mcp.NewToolResultError("Image ID is required"), nil
			}

			artifactID, err := uuid.Parse(artifactIDStr)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse artifact ID", err), nil
			}

			imageID, err := uuid.Parse(imageIDStr)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse image ID", err), nil
			}

			if result, err := m.client.Artifacts().UpdateImage(ctx, artifactID, imageID); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to update Artifact image", err), nil
			} else {
				return JsonToolResult(result)
			}
		},
	}
}
