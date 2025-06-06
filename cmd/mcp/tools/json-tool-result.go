package tools

import (
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
)

func JsonToolResult(result any) (*mcp.CallToolResult, error) {
	if data, err := json.Marshal(result); err != nil {
		return mcp.NewToolResultErrorFromErr("Failed to encode data as JSON", err), nil
	} else {
		return mcp.NewToolResultText(string(data)), nil
	}
}
