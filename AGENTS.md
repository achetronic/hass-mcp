# AGENTS.md

This document helps AI agents work effectively in this repository.

## Project Overview

**MCP Forge** is a production-ready MCP (Model Context Protocol) server template built in Go. It provides a fully compliant MCP server with OAuth authorization support, designed to work with major AI providers (Claude Web, OpenAI) and local clients (Claude Desktop, Cursor, VSCode).

### Key Technologies
- **Language**: Go 1.24+
- **MCP Library**: `github.com/mark3labs/mcp-go` v0.37.0
- **JWT Handling**: `github.com/golang-jwt/jwt/v5`
- **CEL Expressions**: `github.com/google/cel-go` for JWT claim validation
- **Configuration**: YAML with environment variable expansion

## Commands

### Development

```bash
# Run the server (HTTP mode by default)
make run

# Build binary for current platform
make build
# Output: bin/hass-mcp-{os}-{arch}

# Format code
make fmt

# Run go vet
make vet

# Run linter (installs golangci-lint if needed)
make lint

# Run linter with auto-fix
make lint-fix

# Show all available make targets
make help
```

### Docker

```bash
# Build Docker image
make docker-build

# Push Docker image
make docker-push
```

### Packaging

```bash
# Create release package (tar.gz)
make package

# Create MD5 signature for package
make package-signature
```

## Code Organization

```
.
├── cmd/
│   └── main.go              # Application entrypoint
├── api/
│   └── config_types.go      # Configuration type definitions
├── internal/
│   ├── config/
│   │   └── config.go        # YAML config parsing with env expansion
│   ├── globals/
│   │   └── globals.go       # ApplicationContext (config, logger, context)
│   ├── handlers/
│   │   ├── handlers.go      # HandlersManager for HTTP endpoints
│   │   ├── oauth_authorization_server.go  # /.well-known/oauth-authorization-server
│   │   └── oauth_protected_resource.go    # /.well-known/oauth-protected-resource
│   ├── middlewares/
│   │   ├── interfaces.go            # ToolMiddleware, HttpMiddleware interfaces
│   │   ├── jwt_validation.go        # JWT validation middleware
│   │   ├── jwt_validation_utils.go  # JWKS caching, key conversion
│   │   ├── logging.go               # Access logs middleware
│   │   ├── noop.go                  # No-op middleware
│   │   └── utils.go                 # Shared utilities
│   └── tools/
│       ├── tools.go         # ToolsManager - register tools here
│       ├── tool_hello.go    # Example: hello_world tool
│       └── tool_whoami.go   # Example: whoami tool
├── docs/
│   ├── config-http.yaml     # HTTP transport config example
│   └── config-stdio.yaml    # Stdio transport config example
├── chart/                   # Helm chart for Kubernetes deployment
└── .github/workflows/       # CI/CD pipelines
```

## Architecture Patterns

### Dependency Injection Pattern
All components use a consistent pattern with explicit dependencies:

```go
type SomeManagerDependencies struct {
    AppCtx *globals.ApplicationContext
    // ... other dependencies
}

type SomeManager struct {
    dependencies SomeManagerDependencies
}

func NewSomeManager(deps SomeManagerDependencies) *SomeManager {
    return &SomeManager{dependencies: deps}
}
```

### ApplicationContext
Central context object passed throughout the application:
- `Context`: Go context.Context
- `Logger`: slog.Logger (JSON format to stderr)
- `Config`: Parsed YAML configuration

### Tool Registration
Tools are registered in `internal/tools/tools.go`:

```go
func (tm *ToolsManager) AddTools() {
    tool := mcp.NewTool("tool_name",
        mcp.WithDescription("Tool description"),
        mcp.WithString("param",
            mcp.Required(),
            mcp.Description("Parameter description"),
        ),
    )
    tm.dependencies.McpServer.AddTool(tool, tm.HandleToolName)
}
```

Tool handlers are methods on ToolsManager in separate files (`tool_*.go`):

```go
func (tm *ToolsManager) HandleToolName(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Implementation
    return &mcp.CallToolResult{
        Content: []mcp.Content{
            mcp.TextContent{Type: "text", Text: "result"},
        },
    }, nil
}
```

### Middleware Pattern
Two middleware interfaces:
- `HttpMiddleware`: Standard HTTP middleware (`func(next http.Handler) http.Handler`)
- `ToolMiddleware`: MCP tool-level middleware (`func(next server.ToolHandlerFunc) server.ToolHandlerFunc`)

## Configuration

### Environment Variable Expansion
Config files support `$VAR` or `${VAR}` syntax for environment variable expansion.

### Transport Modes
- **HTTP**: For remote clients (Claude Web, OpenAI)
- **Stdio**: For local clients (Claude Desktop, Cursor)

### JWT Validation Strategies
- **external**: JWT validated by upstream proxy (e.g., Istio)
- **local**: JWT validated internally using JWKS URI with CEL expressions for claims

### Key Config Sections
```yaml
server:
  name: "MCP Forge"
  version: "0.1.0"
  transport:
    type: "http"  # or "stdio"
    http:
      host: ":8080"

middleware:
  access_logs:
    excluded_headers: []
    redacted_headers: ["Authorization"]
  jwt:
    enabled: true
    validation:
      strategy: "external"  # or "local"
      forwarded_header: "X-Validated-Jwt"
      local:
        jwks_uri: "https://..."
        cache_interval: "10s"
        allow_conditions:
          - expression: 'payload.groups.exists(g, g == "admin")'

oauth_authorization_server:
  enabled: true
  issuer_uri: "https://keycloak.example.com/realms/mcp-servers"

oauth_protected_resource:
  enabled: true
  resource: "https://hass-mcp.example.com/mcp"
  auth_servers: ["https://keycloak.example.com/realms/mcp-servers"]
```

## Adding New Features

### Adding a New MCP Tool

1. Create handler file `internal/tools/tool_myfeature.go`:
```go
package tools

import (
    "context"
    "github.com/mark3labs/mcp-go/mcp"
)

func (tm *ToolsManager) HandleToolMyFeature(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Get parameters
    args := request.GetArguments()
    param, ok := args["param"].(string)
    if !ok {
        return &mcp.CallToolResult{
            Content: []mcp.Content{
                mcp.TextContent{Type: "text", Text: "Error: param required"},
            },
            IsError: true,
        }, nil
    }
    
    // Return result
    return &mcp.CallToolResult{
        Content: []mcp.Content{
            mcp.TextContent{Type: "text", Text: "Success"},
        },
    }, nil
}
```

2. Register in `internal/tools/tools.go`:
```go
func (tm *ToolsManager) AddTools() {
    // ... existing tools ...
    
    tool = mcp.NewTool("my_feature",
        mcp.WithDescription("Description of what this tool does"),
        mcp.WithString("param",
            mcp.Required(),
            mcp.Description("Parameter description"),
        ),
    )
    tm.dependencies.McpServer.AddTool(tool, tm.HandleToolMyFeature)
}
```

### Adding MCP Resources

1. Copy `internal/tools/` to `internal/resources/`
2. Rename structures (ToolsManager → ResourcesManager, etc.)
3. Wire up in `cmd/main.go` (see commented TODO)

### Adding New HTTP Endpoints

1. Add handler method in `internal/handlers/`:
```go
func (h *HandlersManager) HandleMyEndpoint(response http.ResponseWriter, request *http.Request) {
    // Implementation
}
```

2. Register in `cmd/main.go`:
```go
mux.Handle("/my-endpoint", accessLogsMw.Middleware(http.HandlerFunc(hm.HandleMyEndpoint)))
```

### Adding Configuration Options

1. Add type to `api/config_types.go`
2. Access via `appCtx.Config.YourNewField`

## OAuth Implementation

The server implements:
- **RFC 8414**: OAuth Authorization Server Metadata (`.well-known/oauth-authorization-server`)
- **RFC 9728**: OAuth Protected Resource Metadata (`.well-known/oauth-protected-resource`)

These endpoints support MCP's authorization flow for remote clients like Claude Web.

## Testing

### Manual Testing - Stdio Mode
```bash
make build
# Then configure your MCP client with the binary path and --config flag
```

### Manual Testing - HTTP Mode
```bash
make run
# Server starts on :8080
# Test with curl or mcp-remote package
```

## Deployment

### Kubernetes
Use the Helm chart in `chart/`:
```bash
helm dependency update chart/
helm install hass-mcp chart/
```

The chart uses `bjw-s/app-template` and includes:
- Deployment with configurable replicas
- ConfigMap for configuration
- Service (ClusterIP)
- Optional: ExternalSecret, HTTPRoute, Istio AuthorizationPolicy

### Docker
```bash
make docker-build IMG=your-registry/hass-mcp:tag
make docker-push IMG=your-registry/hass-mcp:tag
```

## Important Gotchas

1. **Logging goes to stderr**: The JSON logger writes to `os.Stderr` to avoid interfering with stdio transport.

2. **JWT header forwarding**: When using external JWT validation (e.g., Istio), the validated JWT is expected in the header specified by `middleware.jwt.validation.forwarded_header`.

3. **CEL expressions**: JWT claim conditions use Google's CEL language. The payload is available as the `payload` variable.

4. **JWKS caching**: In local JWT validation mode, JWKS keys are cached and refreshed at `cache_interval`. A goroutine runs continuously for this.

5. **Config env expansion**: Environment variables in config YAML are expanded at load time via `os.ExpandEnv`.

6. **MCP sessions**: For HTTP transport with sessions, use a consistent hashring proxy (like Hashrouter) in production.

7. **WWW-Authenticate header**: The JWT middleware sets `WWW-Authenticate` header on unauthorized requests per MCP spec, then removes it for authorized requests.

8. **Tool errors vs exceptions**: Return errors in `CallToolResult.Content` with `IsError: true` rather than returning Go errors, unless it's a fatal server error.

## CI/CD

### Release Binaries
On GitHub release creation or manual dispatch:
- Builds for: linux/386, linux/amd64, linux/arm64, darwin/amd64, darwin/arm64
- Uploads tar.gz packages with MD5 signatures

### Release Docker Images
Workflow in `.github/workflows/release-docker-images.yaml` (structure not examined, but exists)
