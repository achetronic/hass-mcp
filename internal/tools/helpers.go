// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"strings"

	"hass-mcp/internal/hass"
)

// Helper functions used across tools

// matchesSearch checks if an entity state matches a search query
func matchesSearch(state hass.EntityState, queryLower string) bool {
	// Search in entity_id
	if strings.Contains(strings.ToLower(state.EntityID), queryLower) {
		return true
	}

	// Search in friendly_name
	if friendlyName, ok := state.Attributes["friendly_name"].(string); ok {
		if strings.Contains(strings.ToLower(friendlyName), queryLower) {
			return true
		}
	}

	// Search in state
	if strings.Contains(strings.ToLower(state.State), queryLower) {
		return true
	}

	return false
}

// entityStateToMap converts an EntityState to a map for JSON output
func entityStateToMap(state hass.EntityState, detailed bool) map[string]interface{} {
	if detailed {
		return map[string]interface{}{
			"entity_id":    state.EntityID,
			"state":        state.State,
			"attributes":   state.Attributes,
			"last_changed": state.LastChanged,
			"last_updated": state.LastUpdated,
		}
	}

	// Lean format
	result := map[string]interface{}{
		"entity_id": state.EntityID,
		"state":     state.State,
	}

	if friendlyName, ok := state.Attributes["friendly_name"]; ok {
		result["friendly_name"] = friendlyName
	}

	// Add domain-specific important attributes
	domain := getDomain(state.EntityID)
	importantAttrs := getDomainImportantAttributes(domain)
	if len(importantAttrs) > 0 {
		attrs := make(map[string]interface{})
		for _, attr := range importantAttrs {
			if val, ok := state.Attributes[attr]; ok {
				attrs[attr] = val
			}
		}
		if len(attrs) > 0 {
			result["attributes"] = attrs
		}
	}

	return result
}

// getEntityName extracts the entity name part from an entity ID
func getEntityName(entityID string) string {
	parts := strings.Split(entityID, ".")
	if len(parts) > 1 {
		return parts[1]
	}
	return entityID
}
