// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// HandleToolGetErrorLog gets the Home Assistant error log
func (tm *ToolsManager) HandleToolGetErrorLog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logText, err := tm.dependencies.HassClient.GetErrorLog()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error getting error log: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Count errors and warnings
	errorCount := strings.Count(logText, "ERROR")
	warningCount := strings.Count(logText, "WARNING")

	// Extract integration mentions
	integrationMentions := make(map[string]int)
	re := regexp.MustCompile(`\[([a-zA-Z0-9_]+)\]`)
	matches := re.FindAllStringSubmatch(logText, -1)
	for _, match := range matches {
		if len(match) > 1 {
			integration := strings.ToLower(match[1])
			integrationMentions[integration]++
		}
	}

	result := map[string]interface{}{
		"log_text":             logText,
		"error_count":          errorCount,
		"warning_count":        warningCount,
		"integration_mentions": integrationMentions,
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error marshaling response: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonBytes),
			},
		},
	}, nil
}
