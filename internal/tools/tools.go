// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"hass-mcp/internal/globals"
	"hass-mcp/internal/hass"
	"hass-mcp/internal/middlewares"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ToolsManagerDependencies struct {
	AppCtx *globals.ApplicationContext

	McpServer   *server.MCPServer
	Middlewares []middlewares.ToolMiddleware
	HassClient  *hass.Client
}

type ToolsManager struct {
	dependencies ToolsManagerDependencies
	toolPrefix   string
}

func NewToolsManager(deps ToolsManagerDependencies) *ToolsManager {
	return &ToolsManager{
		dependencies: deps,
		toolPrefix:   deps.AppCtx.ToolPrefix,
	}
}

func (tm *ToolsManager) toolName(base string) string {
	return tm.toolPrefix + base
}

func (tm *ToolsManager) AddTools() {
	// get_version - Get Home Assistant version
	tool := mcp.NewTool(tm.toolName("get_version"),
		mcp.WithDescription("Get the Home Assistant version"),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.HandleToolGetVersion)

	// get_entity - Get state of a specific entity
	tool = mcp.NewTool(tm.toolName("get_entity"),
		mcp.WithDescription("Get the state of a Home Assistant entity with optional field filtering"),
		mcp.WithString("entity_id",
			mcp.Required(),
			mcp.Description("The entity ID to get (e.g. 'light.living_room')"),
		),
		mcp.WithBoolean("detailed",
			mcp.Description("If true, returns all entity fields without filtering"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.HandleToolGetEntity)

	// entity_action - Perform action on entity (on, off, toggle)
	tool = mcp.NewTool(tm.toolName("entity_action"),
		mcp.WithDescription("Perform an action on a Home Assistant entity (on, off, toggle). Domain-specific params: lights (brightness 0-255, color_temp, rgb_color), covers (position 0-100), climate (temperature, hvac_mode), media_players (source, volume_level 0-1)"),
		mcp.WithString("entity_id",
			mcp.Required(),
			mcp.Description("The entity ID to control (e.g. 'light.living_room')"),
		),
		mcp.WithString("action",
			mcp.Required(),
			mcp.Description("The action to perform: 'on', 'off', or 'toggle'"),
		),
		mcp.WithObject("params",
			mcp.Description("Optional parameters for the service call (e.g. {\"brightness\": 255})"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.HandleToolEntityAction)

	// list_entities - List entities with optional filtering
	tool = mcp.NewTool(tm.toolName("list_entities"),
		mcp.WithDescription("Get a list of Home Assistant entities with optional filtering by domain or search query"),
		mcp.WithString("domain",
			mcp.Description("Optional domain to filter by (e.g., 'light', 'switch', 'sensor')"),
		),
		mcp.WithString("search_query",
			mcp.Description("Optional search term to filter entities by name, id, or attributes"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of entities to return (default: 100)"),
		),
		mcp.WithBoolean("detailed",
			mcp.Description("If true, returns all entity fields without filtering"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.HandleToolListEntities)

	// search_entities - Search for entities
	tool = mcp.NewTool(tm.toolName("search_entities"),
		mcp.WithDescription("Search for entities matching a query string. Returns structured results with domain counts."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The search query to match against entity IDs, names, and attributes"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results to return (default: 20)"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.HandleToolSearchEntities)

	// domain_summary - Get summary of a domain
	tool = mcp.NewTool(tm.toolName("domain_summary"),
		mcp.WithDescription("Get a summary of entities in a specific domain including state distribution, examples, and common attributes"),
		mcp.WithString("domain",
			mcp.Required(),
			mcp.Description("The domain to summarize (e.g., 'light', 'switch', 'sensor')"),
		),
		mcp.WithNumber("example_limit",
			mcp.Description("Maximum number of examples to include for each state (default: 3)"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.HandleToolDomainSummary)

	// system_overview - Get complete system overview
	tool = mcp.NewTool(tm.toolName("system_overview"),
		mcp.WithDescription("Get a comprehensive overview of the entire Home Assistant system including domain counts, samples, and area distribution. Use this as the first call when exploring an unfamiliar instance."),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.HandleToolSystemOverview)

	// list_automations - List all automations
	tool = mcp.NewTool(tm.toolName("list_automations"),
		mcp.WithDescription("Get a list of all automations from Home Assistant including their IDs, states, and friendly names"),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.HandleToolListAutomations)

	// restart_ha - Restart Home Assistant
	tool = mcp.NewTool(tm.toolName("restart_ha"),
		mcp.WithDescription("Restart Home Assistant. WARNING: This will temporarily disrupt all Home Assistant operations."),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.HandleToolRestartHA)

	// call_service - Call any Home Assistant service
	tool = mcp.NewTool(tm.toolName("call_service"),
		mcp.WithDescription("Call any Home Assistant service (low-level API access). Examples: domain='light', service='turn_on', data={'entity_id': 'light.x', 'brightness': 255}"),
		mcp.WithString("domain",
			mcp.Required(),
			mcp.Description("The domain of the service (e.g., 'light', 'switch', 'automation')"),
		),
		mcp.WithString("service",
			mcp.Required(),
			mcp.Description("The service to call (e.g., 'turn_on', 'turn_off', 'toggle', 'reload')"),
		),
		mcp.WithObject("data",
			mcp.Description("Optional data to pass to the service (e.g., {'entity_id': 'light.living_room'})"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.HandleToolCallService)

	// get_history - Get entity history
	tool = mcp.NewTool(tm.toolName("get_history"),
		mcp.WithDescription("Get the history of an entity's state changes. Best for entities with discrete state changes rather than continuously changing sensors."),
		mcp.WithString("entity_id",
			mcp.Required(),
			mcp.Description("The entity ID to get history for"),
		),
		mcp.WithNumber("hours",
			mcp.Description("Number of hours of history to retrieve (default: 24). Keep reasonable (24-72) for efficiency."),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.HandleToolGetHistory)

	// get_error_log - Get Home Assistant error log
	tool = mcp.NewTool(tm.toolName("get_error_log"),
		mcp.WithDescription("Get the Home Assistant error log for troubleshooting. Returns error/warning counts and integration mentions."),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.HandleToolGetErrorLog)
}
