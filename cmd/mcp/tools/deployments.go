package tools

import (
	"context"
	"encoding/json"
	"time"

	"github.com/distr-sh/distr/api"
	"github.com/distr-sh/distr/cmd/mcp/client"
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

			// the put deployment request uses "deploymentId" as the ID field
			data["deploymentId"] = data["id"]

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

// Status tool
func (m *Manager) NewStatusTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"deployment_status",
			mcp.WithString("id", mcp.Required(), mcp.Description("Deployment ID")),
			mcp.WithString("limit", mcp.Description("Limit number of results")),
			mcp.WithString("before", mcp.Description("Before timestamp (RFC3339Nano)")),
			mcp.WithString("after", mcp.Description("After timestamp (RFC3339Nano)")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := ParseUUID(request, "id")
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse deployment ID", err), nil
			}
			var opts client.TimeseriesResourceOptions
			if limit := mcp.ParseInt64(request, "limit", -1); limit != -1 {
				opts.Limit = &limit
			}
			if before := mcp.ParseString(request, "before", ""); before != "" {
				t, err := time.Parse(time.RFC3339Nano, before)
				if err == nil {
					opts.Before = &t
				}
			}
			if after := mcp.ParseString(request, "after", ""); after != "" {
				t, err := time.Parse(time.RFC3339Nano, after)
				if err == nil {
					opts.After = &t
				}
			}
			statuses, err := m.client.Deployments().Status(ctx, id, &opts)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to get deployment status", err), nil
			}
			return JsonToolResult(statuses)
		},
	}
}

// Logs tool
func (m *Manager) NewLogsTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"deployment_logs",
			mcp.WithDescription("Get deployment logs"),
			mcp.WithString("id", mcp.Required(), mcp.Description("Deployment ID")),
			mcp.WithString("resource", mcp.Description("Resource name")),
			mcp.WithString("limit", mcp.Description("Limit number of results")),
			mcp.WithString("before", mcp.Description("Before timestamp (RFC3339Nano)")),
			mcp.WithString("after", mcp.Description("After timestamp (RFC3339Nano)")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := ParseUUID(request, "id")
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse deployment ID", err), nil
			}
			resource := mcp.ParseString(request, "resource", "")
			if resource == "" {
				return mcp.NewToolResultError("resource is required"), nil
			}
			var opts client.TimeseriesResourceOptions
			if limit := mcp.ParseInt64(request, "limit", -1); limit != -1 {
				opts.Limit = &limit
			}
			if before := mcp.ParseString(request, "before", ""); before != "" {
				t, err := time.Parse(time.RFC3339Nano, before)
				if err == nil {
					opts.Before = &t
				}
			}
			if after := mcp.ParseString(request, "after", ""); after != "" {
				t, err := time.Parse(time.RFC3339Nano, after)
				if err == nil {
					opts.After = &t
				}
			}
			logs, err := m.client.Deployments().Logs(ctx, id, resource, &opts)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to get deployment logs", err), nil
			}
			return JsonToolResult(logs)
		},
	}
}

// LogResources tool
func (m *Manager) NewLogResourcesTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"deployment_log_resources",
			mcp.WithDescription("Get available log resources for a deployment"),
			mcp.WithString("id", mcp.Required(), mcp.Description("Deployment ID")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := ParseUUID(request, "id")
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse deployment ID", err), nil
			}
			resources, err := m.client.Deployments().LogResources(ctx, id)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to get log resources", err), nil
			}
			return JsonToolResult(resources)
		},
	}
}
