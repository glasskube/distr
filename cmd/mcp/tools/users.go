package tools

import (
	"context"
	"encoding/json"

	"github.com/glasskube/distr/api"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (m *Manager) NewListUsersTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"list_users",
			mcp.WithDescription("This tool retrieves all user accounts"),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if users, err := m.client.Users().List(ctx); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to list Users", err), nil
			} else {
				return JsonToolResult(users)
			}
		},
	}
}

func (m *Manager) NewCreateUserTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"create_user",
			mcp.WithDescription("This tool creates a new user account"),
			mcp.WithObject("user", mcp.Required(), mcp.Description("User account to create")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var defaultUserData map[string]any
			userData := mcp.ParseStringMap(request, "user", defaultUserData)
			if userData == nil {
				return mcp.NewToolResultError("User data is required"), nil
			}

			userJSON, err := json.Marshal(userData)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to process user data", err), nil
			}

			var user api.CreateUserAccountRequest
			if err := json.Unmarshal(userJSON, &user); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse user data", err), nil
			}

			if result, err := m.client.Users().Create(ctx, user); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to create User", err), nil
			} else {
				return JsonToolResult(result)
			}
		},
	}
}

func (m *Manager) NewDeleteUserTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"delete_user",
			mcp.WithDescription("This tool deletes a user account"),
			mcp.WithString("id", mcp.Required(), mcp.Description("ID of the user to delete")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			idStr := mcp.ParseString(request, "id", "")
			if idStr == "" {
				return mcp.NewToolResultError("ID is required"), nil
			}

			id, err := uuid.Parse(idStr)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse user ID", err), nil
			}

			if err := m.client.Users().Delete(ctx, id); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to delete User", err), nil
			} else {
				return JsonToolResult(map[string]string{"status": "success"})
			}
		},
	}
}

func (m *Manager) NewUpdateUserImageTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"update_user_image",
			mcp.WithDescription("This tool updates a user's image"),
			mcp.WithString("userId", mcp.Required(), mcp.Description("ID of the user")),
			mcp.WithString("imageId", mcp.Required(), mcp.Description("ID of the image")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			userIDStr := mcp.ParseString(request, "userId", "")
			if userIDStr == "" {
				return mcp.NewToolResultError("User ID is required"), nil
			}

			imageIDStr := mcp.ParseString(request, "imageId", "")
			if imageIDStr == "" {
				return mcp.NewToolResultError("Image ID is required"), nil
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse user ID", err), nil
			}

			imageID, err := uuid.Parse(imageIDStr)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse image ID", err), nil
			}

			if result, err := m.client.Users().UpdateImage(ctx, userID, imageID); err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to update User image", err), nil
			} else {
				return JsonToolResult(result)
			}
		},
	}
}
