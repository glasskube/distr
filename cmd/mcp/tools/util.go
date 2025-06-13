package tools

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
)

func ParseT[T any](request mcp.CallToolRequest, key string, defaultValue T) (T, error) {
	var value T
	if mapData := mcp.ParseStringMap(request, key, nil); mapData == nil {
		return defaultValue, nil
	} else if bytesData, err := json.Marshal(mapData); err != nil {
		return value, err
	} else {
		return value, json.Unmarshal(bytesData, &value)
	}
}

func ParseUUID(request mcp.CallToolRequest, key string) (uuid.UUID, error) {
	idStr := mcp.ParseString(request, key, "")
	if idStr == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(idStr)
}
