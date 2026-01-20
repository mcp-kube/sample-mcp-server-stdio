# Sample MCP Server - Stdio Transport

A Model Context Protocol (MCP) server implementation using **stdio (standard input/output)** transport. This server provides utility tools for text processing, currency formatting, and unit conversions.

## Features

### Available Tools

1. **word_count** - Analyze text and count words, characters, and lines
   - Input: `text` (string)
   - Output: Word count, character count, character count without whitespace, line count

2. **format_currency** - Format numbers as currency with proper symbols and decimal places
   - Input: `amount` (number), `currency` (USD, EUR, GBP, or JPY)
   - Output: Formatted currency string (e.g., "$123.45", "¥1234")

3. **slugify** - Convert text to URL-friendly slugs
   - Input: `text` (string)
   - Output: Lowercase, hyphen-separated slug with no special characters

4. **roman_numeral** - Convert between decimal numbers (1-3999) and Roman numerals
   - Input: Either `number` (1-3999) or `roman` (Roman numeral string)
   - Output: Converted value (Roman numeral or decimal number)

5. **temperature_convert** - Convert temperatures between Celsius, Fahrenheit, and Kelvin
   - Input: `value` (number), `from_unit` (celsius/fahrenheit/kelvin), `to_unit` (celsius/fahrenheit/kelvin)
   - Output: Converted temperature value

## Transport: Stdio

This server uses **stdio transport**, which communicates through standard input and output streams. This is ideal for:
- Local command-line tools
- Process-to-process communication
- Shell script integration
- IDE integrations
- Desktop applications

Unlike HTTP-based transports (SSE, Streamable HTTP), stdio provides:
- Lower overhead (no HTTP layer)
- Direct process communication
- Natural fit for CLI tools
- Simpler deployment for local use

## Quick Start

### Using MCP Inspector (Recommended for Testing)

The [MCP Inspector](https://github.com/modelcontextprotocol/inspector) is the official debugging tool for MCP servers:

```bash
# Install the inspector globally
npm install -g @modelcontextprotocol/inspector

# Run the inspector with this server
npx @modelcontextprotocol/inspector go run main.go
```

This will:
1. Start your MCP server as a subprocess
2. Open a web interface at http://localhost:5173
3. Allow you to test all tools interactively

### Using with Claude Desktop

Add this server to your Claude Desktop configuration:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "stdio-tools": {
      "command": "go",
      "args": ["run", "/path/to/sample-mcp-server-stdio/main.go"]
    }
  }
}
```

Or if you've built the binary:

```json
{
  "mcpServers": {
    "stdio-tools": {
      "command": "/path/to/mcp-server-stdio"
    }
  }
}
```

### Building the Binary

```bash
# Build for your current platform
go build -o mcp-server-stdio

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o mcp-server-stdio

# Run the binary
./mcp-server-stdio
```

## Using the Tools

### Example: Word Count

```json
{
  "method": "tools/call",
  "params": {
    "name": "word_count",
    "arguments": {
      "text": "Hello world!\nThis is a test."
    }
  }
}
```

**Response:**
```
Words: 5
Characters: 29
Characters (no whitespace): 24
Lines: 2
```

### Example: Format Currency

```json
{
  "method": "tools/call",
  "params": {
    "name": "format_currency",
    "arguments": {
      "amount": 1234.56,
      "currency": "USD"
    }
  }
}
```

**Response:** `$1234.56`

### Example: Slugify

```json
{
  "method": "tools/call",
  "params": {
    "name": "slugify",
    "arguments": {
      "text": "Hello World! This is a Test."
    }
  }
}
```

**Response:** `hello-world-this-is-a-test`

### Example: Roman Numeral Conversion

Convert to Roman:
```json
{
  "method": "tools/call",
  "params": {
    "name": "roman_numeral",
    "arguments": {
      "number": 2024
    }
  }
}
```

**Response:** `MMXXIV`

Convert from Roman:
```json
{
  "method": "tools/call",
  "params": {
    "name": "roman_numeral",
    "arguments": {
      "roman": "MMXXIV"
    }
  }
}
```

**Response:** `2024`

### Example: Temperature Conversion

```json
{
  "method": "tools/call",
  "params": {
    "name": "temperature_convert",
    "arguments": {
      "value": 100,
      "from_unit": "celsius",
      "to_unit": "fahrenheit"
    }
  }
}
```

**Response:** `212.00`

## Client Library Examples

### TypeScript/JavaScript

```typescript
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";

const transport = new StdioClientTransport({
  command: "./mcp-server-stdio",
  args: []
});

const client = new Client({
  name: "example-client",
  version: "1.0.0"
}, {
  capabilities: {}
});

await client.connect(transport);

// Use the word_count tool
const result = await client.callTool({
  name: "word_count",
  arguments: {
    text: "Hello world! This is a test."
  }
});

console.log(result);
```

### Python

```python
from mcp.client.stdio import stdio_client, StdioServerParameters
from mcp.types import CallToolRequest

server_params = StdioServerParameters(
    command="./mcp-server-stdio",
    args=[]
)

async with stdio_client(server_params) as (read, write):
    # Initialize the connection
    await write({"jsonrpc": "2.0", "method": "initialize", "params": {...}})

    # Call a tool
    result = await write(CallToolRequest(
        method="tools/call",
        params={
            "name": "slugify",
            "arguments": {"text": "Hello World!"}
        }
    ))

    print(result)
```

## Docker Deployment

### Build the Docker Image

```bash
docker build -t sample-mcp-server-stdio:latest .
```

### Run with Docker

Since stdio requires interactive terminal access:

```bash
docker run -i sample-mcp-server-stdio:latest
```

Note: Stdio servers in Docker are primarily useful for:
- Building and testing the container
- Using as a base for sidecar containers
- Integration into larger container orchestration systems

For typical MCP server deployments, consider HTTP-based transports (SSE or Streamable HTTP).

## Kubernetes Deployment

While stdio transport can run in Kubernetes, it's less common than HTTP-based transports. The provided manifests are included for completeness:

```bash
# Deploy to Kubernetes
kubectl apply -k k8s/

# Check deployment status
kubectl get pods -l app=sample-mcp-server-stdio

# Note: Accessing stdio servers in K8s requires kubectl exec or similar
kubectl exec -it deployment/sample-mcp-server-stdio -- /bin/sh
```

For production Kubernetes deployments, consider using the SSE or Streamable HTTP variants instead.

## Development

### Project Structure

```
sample-mcp-server-stdio/
├── main.go              # Main server implementation
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
├── Dockerfile           # Container build configuration
├── .dockerignore        # Docker build exclusions
├── README.md            # This file
├── .claude/             # Claude IDE settings
│   └── settings.local.json
└── k8s/                 # Kubernetes manifests
    ├── deployment.yaml
    ├── service.yaml
    └── kustomization.yaml
```

### Adding New Tools

To add a new tool:

1. Define the argument struct with JSON schema tags:
```go
type MyToolArgs struct {
    Input string `json:"input" jsonschema:"description=Input description"`
}
```

2. Create the handler function:
```go
func handleMyTool(ctx context.Context, req *mcp.CallToolRequest, args MyToolArgs) (*mcp.CallToolResult, any, error) {
    // Your tool logic here
    return &mcp.CallToolResult{
        Content: []any{
            mcp.TextContent{
                Type: "text",
                Text: "result",
            },
        },
    }, nil, nil
}
```

3. Register the tool in `main()`:
```go
mcp.AddTool(server, "my_tool", "Tool description",
    jsonschema.For[MyToolArgs](),
    handleMyTool)
```

## Logging

All logs are written to **stderr** to keep stdout clean for MCP protocol messages. Log format:

```
[YYYY-MM-DD HH:MM:SS.mmmmmm] [TAG] message
```

Log tags:
- `[MAIN]` - Server lifecycle events
- `[TOOL]` - Tool execution
- `[ERROR]` - Error messages

## Dependencies

- **Go SDK**: `github.com/modelcontextprotocol/go-sdk` v1.2.0
- **JSON Schema**: `github.com/google/jsonschema-go` v0.3.0

## Related Implementations

- **sample-mcp-server-sse**: Server using Server-Sent Events (SSE) transport
- **sample-mcp-server-streamable-http**: Server using Streamable HTTP transport

Each transport has different use cases:
- **Stdio**: Local tools, CLI integration, desktop apps
- **SSE**: Real-time streaming, web applications, push notifications
- **Streamable HTTP**: Traditional request/response, REST-like APIs

## License

This is a sample implementation for educational purposes.

## Learn More

- [Model Context Protocol Documentation](https://modelcontextprotocol.io)
- [MCP Specification](https://spec.modelcontextprotocol.io)
- [MCP TypeScript SDK](https://github.com/modelcontextprotocol/typescript-sdk)
- [MCP Python SDK](https://github.com/modelcontextprotocol/python-sdk)
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
