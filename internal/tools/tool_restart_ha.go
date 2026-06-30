// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// HandleToolRestartHA restarts Home Assistant
func (tm *ToolsManager) HandleToolRestartHA(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	err := tm.dependencies.HassClient.RestartHomeAssistant()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error restarting Home Assistant: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	result := map[string]interface{}{
		"success": true,
		"message": "Home Assistant restart initiated. The system will be temporarily unavailable.",
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonBytes),
			},
		},
	}, nil
}
