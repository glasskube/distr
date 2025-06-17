package tools

import (
	"context"
	"encoding/json"

	"github.com/glasskube/distr/api"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (m *Manager) NewCreateOrUpdateDeploymentTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"create_or_update_deployment",
			mcp.WithDescription("This tool creates or updates a deployment"),
			mcp.WithObject("deployment", mcp.Required(), mcp.Description("Deployment request object")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			data := mcp.ParseStringMap(request, "deployment", nil)
			if data == nil {
				return mcp.NewToolResultError("deployment object is required"), nil
			}

			dataJSON, err := json.Marshal(data)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to process deployment data", err), nil
			}

			var deployment api.DeploymentRequest
			if err := json.Unmarshal(dataJSON, &deployment); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse deployment object", err), nil
			}

			if err := m.client.Deployments().Put(ctx, deployment); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to create or update Deployment", err), nil
			} else {
				return JsonToolResult(map[string]string{"status": "success"})
			}
		},
	}
}

func (m *Manager) NewPatchDeploymentTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"patch_deployment",
			mcp.WithDescription("This tool patches an existing deployment"),
			mcp.WithString("id", mcp.Required(), mcp.Description("ID of the deployment to patch")),
			mcp.WithObject("patch", mcp.Required(), mcp.Description("Patch request object")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := ParseUUID(request, "id")
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse deployment ID", err), nil
			}
			if id == uuid.Nil {
				return mcp.NewToolResultError("id is required"), nil
			}

			patchData := mcp.ParseStringMap(request, "patch", nil)
			if patchData == nil {
				return mcp.NewToolResultError("patch object is required"), nil
			}

			dataJSON, err := json.Marshal(patchData)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to process patch data", err), nil
			}

			var patch api.PatchDeploymentRequest
			if err := json.Unmarshal(dataJSON, &patch); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse patch object", err), nil
			}

			if result, err := m.client.Deployments().Patch(ctx, id, patch); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to patch Deployment", err), nil
			} else {
				return JsonToolResult(result)
			}
		},
	}
}

func (m *Manager) NewDeleteDeploymentTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"delete_deployment",
			mcp.WithDescription("This tool deletes a deployment"),
			mcp.WithString("id", mcp.Required(), mcp.Description("ID of the deployment to delete")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			idStr := mcp.ParseString(request, "id", "")
			if idStr == "" {
				return mcp.NewToolResultError("id is required"), nil
			}

			id, err := uuid.Parse(idStr)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse deployment ID", err), nil
			}

			if err := m.client.Deployments().Delete(ctx, id); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to delete Deployment", err), nil
			} else {
				return JsonToolResult(map[string]string{"status": "success"})
			}
		},
	}
}
