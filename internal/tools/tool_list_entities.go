// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// HandleToolListEntities lists Home Assistant entities with optional filtering
func (tm *ToolsManager) HandleToolListEntities(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	// Get optional domain filter
	domain, _ := args["domain"].(string)

	// Get optional search query
	searchQuery, _ := args["search_query"].(string)

	// Get limit (default 100)
	limit := 100
	if limitVal, ok := args["limit"].(float64); ok {
		limit = int(limitVal)
	}

	// Get detailed flag
	detailed, _ := args["detailed"].(bool)

	var entities []map[string]interface{}
	var err error

	if domain != "" {
		// Filter by domain
		states, err := tm.dependencies.HassClient.GetEntitiesByDomain(domain)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Error getting entities: %v", err),
					},
				},
				IsError: true,
			}, nil
		}

		// Apply search filter if provided
		if searchQuery != "" && searchQuery != "*" {
			queryLower := strings.ToLower(searchQuery)
			var filtered []map[string]interface{}
			for _, state := range states {
				if matchesSearch(state, queryLower) {
					filtered = append(filtered, entityStateToMap(state, detailed))
				}
			}
			entities = filtered
		} else {
			for _, state := range states {
				entities = append(entities, entityStateToMap(state, detailed))
			}
		}
	} else if searchQuery != "" && searchQuery != "*" {
		// Search across all entities
		states, err := tm.dependencies.HassClient.SearchEntities(searchQuery, limit)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Error searching entities: %v", err),
					},
				},
				IsError: true,
			}, nil
		}
		for _, state := range states {
			entities = append(entities, entityStateToMap(state, detailed))
		}
	} else {
		// Get all entities
		states, err := tm.dependencies.HassClient.GetAllStates()
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Error getting entities: %v", err),
					},
				},
				IsError: true,
			}, nil
		}
		for _, state := range states {
			entities = append(entities, entityStateToMap(state, detailed))
		}
	}

	// Apply limit
	if limit > 0 && len(entities) > limit {
		entities = entities[:limit]
	}

	_ = err // silence unused variable if search branch not taken

	jsonBytes, err := json.MarshalIndent(entities, "", "  ")
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
