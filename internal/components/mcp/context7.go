package mcp

import "github.com/ismartz/aispace-setup/internal/model"

// claudeSeparateMCPConfig returns the JSON content for a single MCP server config file.
// Claude Code uses one file per server at ~/.claude/mcp/<server>.json
// Format: {"mcpServers": {"<name>": {...}}}
func claudeSeparateMCPConfig(serverName string) []byte {
	switch serverName {
	case "context7":
		return []byte(`{
  "mcpServers": {
    "context7": {
      "type": "http",
      "url": "https://mcp.context7.com/mcp"
    }
  }
}
`)
	case "aispace":
		return []byte(`{
  "mcpServers": {
    "aispace": {
      "type": "stdio",
      "command": "node",
      "args": ["${HOME}/.aispace-setup/mcp/index.js"]
    }
  }
}
`)
	default:
		return nil
	}
}

// opencodeMCPOverlay returns the JSON overlay to merge into opencode.json for MCP config.
// OpenCode uses a single "mcp" key with all servers merged together.
func opencodeMCPOverlay() []byte {
	return []byte(`{
  "mcp": {
    "context7": {
      "type": "http",
      "url": "https://mcp.context7.com/mcp"
    },
    "aispace": {
      "type": "stdio",
      "command": "node",
      "args": ["${HOME}/.aispace-setup/mcp/index.js"]
    }
  }
}
`)
}

// mcpServersForAgent returns the list of MCP server names to configure for an agent.
func mcpServersForAgent(agent model.AgentID) []string {
	return []string{"context7", "aispace"}
}
