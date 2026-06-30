// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// HandleToolGetVersion returns the Home Assistant version
func (tm *ToolsManager) HandleToolGetVersion(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	version, err := tm.dependencies.HassClient.GetVersion()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error getting Home Assistant version: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: version,
			},
		},
	}, nil
}
