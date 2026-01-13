# Home Assistant MCP Server

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/achetronic/hass-mcp)
![GitHub](https://img.shields.io/github/license/achetronic/hass-mcp)

An MCP server that exposes Home Assistant functionality to AI assistants. Built in Go, it enables AI clients like Claude, Cursor, or OpenAI to interact with your smart home.

## Overview

This server implements the Model Context Protocol to bridge AI assistants with Home Assistant. It provides tools for querying entity states, controlling devices, viewing automations, and troubleshooting your Home Assistant installation.

## Features

- **Entity Management**: Query, search, and control Home Assistant entities
- **Service Calls**: Execute any Home Assistant service through the AI
- **System Insights**: Get overviews, domain summaries, and entity history
- **Troubleshooting**: Access error logs and automation states
- **Flexible Transport**: Supports both stdio (local) and HTTP (remote) modes
- **Production Ready**: Includes OAuth support, JWT validation, Dockerfile, and Helm chart

## Available Tools

| Tool               | Description                                       |
|--------------------|---------------------------------------------------|
| `get_version`      | Get the Home Assistant version                    |
| `get_entity`       | Get state of a specific entity                    |
| `entity_action`    | Turn entities on, off, or toggle them             |
| `list_entities`    | List entities with optional domain/search filters |
| `search_entities`  | Search entities by name, ID, or attributes        |
| `domain_summary`   | Get statistics for a specific domain              |
| `system_overview`  | Get a complete overview of the HA system          |
| `list_automations` | List all automations with their states            |
| `call_service`     | Call any Home Assistant service                   |
| `get_history`      | Get state change history for an entity            |
| `get_error_log`    | Retrieve the Home Assistant error log             |
| `restart_ha`       | Restart Home Assistant                            |

## Requirements

- Go 1.24 or later
- Home Assistant instance with API access
- Long-lived access token from Home Assistant

## Configuration

Create a configuration file based on the examples in `docs/`:

```yaml
server:
  name: "Home Assistant MCP"
  version: "0.1.0"
  transport:
    type: "stdio"  # or "http"

home_assistant:
  url: "http://homeassistant.local:8123"
  token: "your-long-lived-access-token"
```

Environment variables are supported using `${VAR}` syntax:

```yaml
home_assistant:
  url: "${HA_URL}"
  token: "${HA_TOKEN}"
```

### Obtaining a Home Assistant Token

1. Navigate to your Home Assistant instance
2. Click on your profile (bottom left)
3. Scroll to "Long-Lived Access Tokens"
4. Create a new token and copy it

## Installation

### Build from Source

```bash
make build
```

The binary will be created at `bin/hass-mcp-{os}-{arch}`.

### Docker

```bash
make docker-build IMG=your-registry/hass-mcp:latest
```

## Usage

### Local Clients (Claude Desktop, Cursor)

For local AI clients, use stdio transport. Add to your client configuration:

```json
{
  "mcpServers": {
    "home-assistant": {
      "command": "/path/to/hass-mcp",
      "args": ["--config", "/path/to/config-stdio.yaml"],
      "env": {
        "HA_URL": "http://homeassistant.local:8123",
        "HA_TOKEN": "your-token-here"
      }
    }
  }
}
```

### Remote Clients (Claude Web, OpenAI)

For remote clients, use HTTP transport:

```bash
make run
```

This starts the server on port 8080 by default. The server exposes:
- `/mcp` - MCP protocol endpoint
- `/.well-known/oauth-authorization-server` - OAuth metadata (if enabled)
- `/.well-known/oauth-protected-resource` - Protected resource metadata (if enabled)

## Development

```bash
# Run the server
make run

# Format code
make fmt

# Run linter
make lint

# Build binary
make build
```

## Deployment

### Kubernetes

A Helm chart is provided in the `chart/` directory:

```bash
helm install hass-mcp ./chart
```

Configure values in `chart/values.yaml` to match your environment.

### Production Recommendations

- Use a reverse proxy (e.g., Istio) for JWT validation
- Enable OAuth endpoints for remote client authentication
- Use a consistent hash router for session affinity if running multiple replicas

## Project Structure

```
├── cmd/main.go              # Application entry point
├── api/                     # Configuration types
├── internal/
│   ├── hass/                # Home Assistant API client
│   ├── tools/               # MCP tool implementations
│   ├── handlers/            # HTTP handlers
│   ├── middlewares/         # HTTP middlewares
│   ├── config/              # Configuration loading
│   └── globals/             # Application context
├── docs/                    # Example configurations
└── chart/                   # Helm chart
```

## Documentation

- [Model Context Protocol Specification](https://modelcontextprotocol.io/)
- [Home Assistant REST API](https://developers.home-assistant.io/docs/api/rest/)
- [mcp-go Library](https://github.com/mark3labs/mcp-go)

## Contributing

Contributions are welcome. Please open an issue to discuss changes before submitting a pull request.

## License

This project is licensed under the Apache 2.0 License. See [LICENSE](./LICENSE) for details.
