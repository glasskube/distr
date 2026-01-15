package tools

import (
	"context"

	"github.com/distr-sh/distr/api"
	"github.com/distr-sh/distr/cmd/mcp/client"
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
			user, err := ParseT[*api.CreateUserAccountRequest](request, "user", nil)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse user data", err), nil
			} else if user == nil {
				return mcp.NewToolResultError("User data is required"), nil
			}

			if result, err := m.client.Users().Create(ctx, *user); err != nil {
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
			id, err := ParseUUID(request, "id")
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Failed to parse user ID", err), nil
			}
			if id == uuid.Nil {
				return mcp.NewToolResultError("ID is required"), nil
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
		Handler: m.patchImageHandlerFunc(
			"userId",
			"imageId",
			func(c *client.Client) func(context.Context, uuid.UUID, uuid.UUID) (any, error) {
				return func(ctx context.Context, u1, u2 uuid.UUID) (any, error) {
					if result, err := c.Users().UpdateImage(ctx, u1, u2); err != nil {
						return nil, err
					} else {
						return result, nil
					}
				}
			},
			"Failed to update User image"),
	}
}
