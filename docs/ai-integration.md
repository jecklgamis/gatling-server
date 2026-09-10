# AI Integration

[gatling-mcp-server](https://github.com/jecklgamis/gatling-mcp-server) wraps this API as an MCP (Model Context
Protocol) server, so a simulation can be uploaded, submitted, monitored, and aborted just by describing what you
want in plain English instead of hand-writing `curl` calls. It can be configured in any MCP-capable AI client -
Claude Code, Claude Desktop, Cursor, Windsurf, Cline, Gemini CLI, JetBrains AI Assistant, and others - by pointing
it at a running gatling-mcp-server instance.

See gatling-mcp-server's own docs for:

- **[Connecting Clients](https://jecklgamis.github.io/gatling-mcp-server/#/clients)** - per-client setup (Claude
  Code, Claude Desktop, Cursor, Windsurf, Cline, Gemini CLI, JetBrains AI Assistant).
- **[Architecture](https://jecklgamis.github.io/gatling-mcp-server/#/architecture)** - how gatling-mcp-server,
  gatling-server, and the example Gatling projects fit together, including the local-vs-remote deployment
  tradeoff and the recommended split between scripted jar publishing and agent-driven task operation.
