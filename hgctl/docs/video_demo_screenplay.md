# Hgctl Agent Module - Competition Defense Video Demo Screenplay

## Video Structure Overview
**Total Duration:** 10-12 minutes
**Target Audience:** Competition judges and technical evaluators
**Goal:** Demonstrate complete workflow from zero to production-ready DeepResearch Agent API Product

---

## Act 1: Introduction & Value Proposition (1-2 min)

### Scene 1.1: Opening Hook (15 seconds)
**Visual:** Terminal screen with hgctl logo/banner
**Narration:**
> "Building production-grade AI Agents today is painful. Integration hell with countless APIs, disconnected tooling, and the massive gap between prototype and production. What if you could go from idea to deployed Agent API in minutes, not weeks?"

### Scene 1.2: Problem Statement (30 seconds)
**Visual:** Split screen showing traditional agent development pain points
**Narration:**
> "Current Agent development faces three major challenges:
> - First, **Integration Hell** - connecting agents to existing APIs requires massive custom development
> - Second, **Fragmented Experience** - agent logic, tool development, and infrastructure management are completely disconnected
> - Third, **Production Gap** - most agent frameworks lack security, observability, and governance needed for real production use"

### Scene 1.3: Solution Overview (30 seconds)
**Visual:** Architecture diagram showing Hgctl + Higress + Himarket integration
**Narration:**
> "Introducing Hgctl Agent Module - a comprehensive solution that transforms AI agent development:
> - Uses **MCP Protocol** to standardize tool integration
> - Integrates powerful **AgentCore** tools like Claude Code as the interactive engine
> - Leverages **Higress AI Gateway** for production-grade routing, auth, and governance
> - Publishes to **Himarket** for API marketplace capabilities
>
> Today, I'll demonstrate by building a DeepResearch Agent that can analyze GitHub repositories and generate comprehensive research reports."

---

## Act 2: Environment Setup & Configuration (1.5-2 min)

### Scene 2.1: Configuration Display (30 seconds)
**Command:**
```bash
cat ~/.hgctl.json
```

**Expected Output:**
```json
{
  "hgctl-agent-core": "claude",
  "higress-console-url": "http://localhost:8080",
  "higress-console-user": "admin",
  "higress-console-password": "admin123",
  "higress-gateway-url": "http://10.96.0.1:80",
  "himarket-admin-url": "http://localhost:5174",
  "himarket-admin-user": "admin",
  "himarket-admin-password": "admin123"
}
```

**Narration:**
> "Hgctl uses a flexible multi-layer configuration system. You can configure via JSON files, environment variables, or command-line flags. This allows seamless switching between development, testing, and production environments. Here we've configured Claude Code as our AgentCore, along with Higress Console and Himarket credentials."

**Additional Visual:** Show environment variables as alternative
```bash
export HGCTL_AGENT_CORE=claude
export HIGRESS_CONSOLE_URL=http://localhost:8080
```

### Scene 2.2: Environment Auto-Initialization (45 seconds)
**Command:**
```bash
hgctl agent
```

**Expected Output/Process:**
```
🔍 Checking AgentCore availability...
✅ Found claude-code (v1.2.0)

🔍 Checking Node.js environment...
✅ Node.js v20.10.0 detected

🔍 Checking Python environment...
✅ Python 3.11.5 detected

🔄 Initializing agent workspace...
📁 Creating ~/.hgctl/agents/
📁 Creating ~/.hgctl/templates/
✅ Workspace initialized

🔌 Configuring built-in MCP servers...
  ✅ higress-api MCP server registered
  ✅ deepwiki MCP server available

🎉 Initialization complete!
Starting interactive agent window...
```

**Narration:**
> "First-time setup is fully automated. Hgctl detects and installs required environments - Node.js 18+, Python 3.10+, and the selected AgentCore. It initializes the workspace and pre-configures essential MCP servers, including the Higress-API server that lets agents manage infrastructure through natural language. Let me start the interactive window."

**Visual:** Show the Claude Code interactive interface starting up

### Scene 2.3: Interactive Agent Window Demo (30 seconds)
**Action:** In the agent window, type a simple command to demonstrate the Higress-API MCP server

**User Input in Agent Window:**
```
Can you list all the MCP servers currently registered in Higress?
```

**Expected Agent Response:**
```
I'll query the Higress Console API to list MCP servers...

[Uses higress-api MCP server]

Found 3 MCP servers:
1. deepwiki - GitHub repository documentation tool
2. weather-api - OpenAPI weather service
3. higress-api - Higress management API

Would you like more details about any of these?
```

**Narration:**
> "The interactive window is now running with Claude Code. Notice how the agent can directly call Higress management APIs through the pre-configured MCP server - this is the foundation for infrastructure-as-code through natural language. Let me exit and show you the MCP management capabilities."

**Command to exit:**
```
exit
```

---

## Act 3: MCP Server Management (2-3 min)

### Scene 3.1: HTTP MCP Server - DeepWiki (1 min)
**Narration:**
> "For our DeepResearch Agent, we'll need the DeepWiki MCP server - a tool that fetches and analyzes GitHub repository documentation. Since DeepWiki already exposes an MCP-compatible endpoint, we can use HTTP type for direct integration."

**Command:**
```bash
hgctl mcp add deepwiki https://mcp-deepwiki.dns/mcp \
  --type http \
  --transport streamable \
  --description "GitHub repository deep research and documentation analysis tool"
```

**Expected Output:**
```
🔍 Validating MCP endpoint...
✅ MCP endpoint is accessible and valid

📋 MCP Server Configuration:
  Name: deepwiki
  Type: HTTP
  Transport: streamable
  URL: https://mcp-deepwiki.dns/mcp

🚀 Publishing to Higress Console...
  ✅ Created service source: deepwiki.dns:443
  ✅ Created MCP server route: /mcp-servers/deepwiki

🎯 Registering to AgentCore (claude)...
  ✅ Added to ~/.claude/mcp.json

✅ MCP Server 'deepwiki' is ready to use!

📍 Access URL: http://10.96.0.1:80/mcp-servers/deepwiki
```

**Narration:**
> "The command validates the endpoint, publishes it to Higress Console creating the necessary service and route configurations, and registers it to Claude Code's MCP config. Now this tool is available both in the interactive window and can be shared cluster-wide."

**Action:** Verify the MCP server in Higress Console
**Command:**
```bash
curl -s http://localhost:8080/v1/mcpServer | jq '.data[] | select(.name=="deepwiki")'
```

**Expected Output (beautified):**
```json
{
  "name": "deepwiki",
  "description": "GitHub repository deep research and documentation analysis tool",
  "type": "DIRECT_ROUTE",
  "services": [
    {
      "name": "mcp-deepwiki.dns",
      "port": 443,
      "weight": 100
    }
  ],
  "directRouteConfig": {
    "path": "/mcp",
    "transportType": "streamable"
  }
}
```

### Scene 3.2: OpenAPI MCP Server - ArXiv Search (1.5 min)
**Narration:**
> "For research agents, we also want to search academic papers. We have an ArXiv search API with an OpenAPI spec, but it doesn't natively support MCP. Hgctl can automatically convert any OpenAPI spec into an MCP server."

**Action:** Show the OpenAPI spec file (quick glimpse)
```bash
cat ~/arxiv-api-spec.yaml | head -20
```

**Expected Output:**
```yaml
openapi: 3.0.0
info:
  title: ArXiv Search API
  version: 1.0.0
  description: Search and retrieve ArXiv academic papers
servers:
  - url: https://export.arxiv.org/api
paths:
  /query:
    get:
      summary: Search papers
      operationId: searchPapers
      parameters:
        - name: search_query
          in: query
          required: true
          schema:
            type: string
...
```

**Command:**
```bash
hgctl mcp add arxiv-search ~/arxiv-api-spec.yaml \
  --type openapi \
  --description "Academic paper search from ArXiv"
```

**Expected Output:**
```
📄 Parsing OpenAPI specification...
  ✅ Found 5 endpoints
  ✅ Extracted 3 schemas
  ✅ Validated specification

🔄 Converting to MCP format...
  ✅ Generated MCP tools configuration
  ✅ Created tool: searchPapers
  ✅ Created tool: getPaperDetails
  ✅ Created tool: listCategories

📋 MCP Configuration Preview:
  Server Name: arxiv-search
  Tools: 5
  Transport: streamable_http

🚀 Publishing to Higress Console...
  ✅ Created service source: export.arxiv.org:443
  ✅ Created MCP Server: arxiv-search
  ✅ Created OpenAPI tool configuration

🌐 Getting Higress Gateway cluster IP...
  ✅ Gateway IP: 10.96.0.1

🎯 Registering to AgentCore...
  ✅ Added to ~/.claude/mcp.json
  MCP URL: http://10.96.0.1:80/mcp-servers/arxiv-search

✅ OpenAPI → MCP conversion complete!
```

**Narration:**
> "Hgctl automatically parsed the OpenAPI spec, extracted all endpoints and schemas, converted them to MCP tool definitions, and published everything to Higress. The OpenAPI-to-MCP conversion happens on the Higress side using the MCP plugin, so there's zero additional deployment needed. The agent can now search academic papers through MCP just like any other tool."

### Scene 3.3: Publishing MCP Server as API Product to Himarket (45 seconds)
**Narration:**
> "Let's make the DeepWiki tool available as an API Product in Himarket, so other developers can discover and use it."

**Command:**
```bash
hgctl mcp add deepwiki-product https://mcp-deepwiki.dns/mcp \
  --type http \
  --transport streamable \
  --as-product \
  --description "DeepWiki - GitHub repository analysis as a service"
```

**Expected Output:**
```
🔍 Validating MCP endpoint...
✅ MCP endpoint is accessible

🚀 Publishing to Higress Console...
  ✅ MCP Server created: deepwiki-product

📦 Publishing to Himarket as API Product...
  🏷️  Product Type: MCP_SERVER
  📝 Creating product...
  ✅ Product created (ID: prod-mcp-deepwiki-20250120)
  🔗 Linking to Higress gateway...
  ✅ Gateway reference established

✅ MCP Server 'deepwiki-product' is now available in Himarket!

📍 Himarket URL: http://localhost:5174/products/mcp-server/deepwiki-product
```

**Visual:** Switch to browser showing Himarket portal with the new MCP Server product listed

**Narration:**
> "The MCP server is now listed in Himarket. Developers can browse available tools, see documentation auto-generated from the MCP schema, apply for access, and monitor usage - all through the unified API marketplace."

---

## Act 4: Creating the DeepResearch Agent (3-4 min)

### Scene 4.1: Initiating Agent Creation (20 seconds)
**Command:**
```bash
hgctl agent new
```

**Narration:**
> "Now let's create our DeepResearch Agent. The 'agent new' command starts an interactive wizard that guides you through the entire configuration process."

### Scene 4.2: Interactive Agent Configuration (2.5 min)

**Initial Prompt:**
```
? How would you like to create an agent
  ❯ create step by step
    import existing one from current agentcore (claude)
```
**Action:** Select "create step by step"

**Narration:**
> "We have two options: create step-by-step with guided prompts, or import an existing agent definition from Claude Code. For this demo, let's create from scratch to show all configuration options."

---

**Step 1: Agent Name**
```
? What is the agent's name? deepresearch
```

**Narration:**
> "First, we name our agent. This will be used as the service identifier and directory name."

---

**Step 2: App Description**
```
🤖 Let's configure your agent 'deepresearch'

📋 App Description
  A brief description of what this agent does
? What is the app description?
> Analyzes GitHub repositories and generates comprehensive research reports combining code analysis, documentation, and related academic papers
```

---

**Step 3: System Prompt**
```
📝 System Prompt
  This defines the agent's personality and behavior
? How would you like to set the agent's SysPrompt
  ❯ input directly
    use existing markdown file
    use LLM to generate
```
**Action:** Select "use LLM to generate"

**Narration:**
> "For the system prompt, we have three options: direct input, use an existing markdown file, or let the LLM generate one based on a description. I'll use LLM generation to show this powerful feature."

```
? Describe what this agent should do (be comprehensive for best results)
> This agent is a deep research assistant specialized in analyzing GitHub repositories. It should:
  1. Clone and analyze repository structure, key files, and code patterns
  2. Extract and summarize documentation (README, wikis, docs folder)
  3. Understand the project's tech stack and dependencies
  4. Search for related academic papers on ArXiv
  5. Identify key contributors and development patterns
  6. Generate a comprehensive research report in markdown format
  Be professional, thorough, and cite all sources.

⏳ generating...(this may take a few minutes, depends on your model)
```

**Expected Generated Prompt:**
```
Generate Prompt for agent deepresearch:

# DeepResearch Agent - System Prompt

You are DeepResearch, an expert AI research assistant specialized in comprehensive GitHub repository analysis and technical research synthesis.

## Core Capabilities

1. **Repository Analysis**
   - Examine repository structure, architecture patterns, and code organization
   - Identify key files, entry points, and critical components
   - Analyze dependencies and technology stack
   - Evaluate code quality, testing practices, and documentation

2. **Documentation Synthesis**
   - Extract and summarize README files, wikis, and documentation folders
   - Understand project goals, features, and usage patterns
   - Identify installation procedures, API references, and examples

3. **Academic Context**
   - Search ArXiv for related academic papers and research
   - Connect repository implementations to theoretical foundations
   - Cite relevant papers that inform the project's approach

4. **Development Insights**
   - Analyze commit history and contributor patterns
   - Identify development velocity and project health
   - Recognize coding conventions and best practices

## Output Format

Generate research reports in structured markdown with these sections:
- Executive Summary
- Project Overview (purpose, tech stack, architecture)
- Key Features & Implementation Details
- Documentation Quality Assessment
- Related Academic Research
- Development Activity & Community Health
- Recommendations & Insights
- References

## Guidelines

- Be thorough yet concise
- Always cite sources (repository files, documentation URLs, paper references)
- Use clear technical language
- Highlight both strengths and areas for improvement
- Provide actionable insights
```

**Narration:**
> "The LLM generated a comprehensive system prompt that defines the agent's role, capabilities, output format, and guidelines. This is saved as a markdown file for easy editing later."

---

**Step 4: AI Provider & Model**
```
🏢 AI Provider
? Choose the AI provider (DashScope):
    DashScope
  ❯ OpenAI
    Anthropic
    Ollama
    Gemini
    Trinity
```
**Action:** Select "Anthropic"

```
🤖 AI Model
? Which model to use? (claude-3-5-sonnet-latest)
> claude-3-5-sonnet-20241022
```

```
🔑 API Key Configuration
? Environment variable name for API key: (ANTHROPIC_API_KEY)
> ANTHROPIC_API_KEY
```

**Narration:**
> "We select Anthropic's Claude as the reasoning engine, with the API key securely stored in an environment variable. Hgctl supports all major AI providers out of the box."

---

**Step 5: Tools Selection**
```
🔧 Available Tools
  Select the tools this agent can use
   • PythonCode
   • ExecuteShellCommand
   • ReadTextFile
   • ReadImageFile
   • WriteFile
   • CreateDirectory
   • MoveFile
   • ImageGeneration
   • TextToImage

? Which tools to enable? (Space to select, Enter to confirm)
  ❯ ◉ PythonCode
    ◉ ExecuteShellCommand
    ◉ ReadTextFile
    ◉ WriteFile
    ◯ ReadImageFile
    ◯ CreateDirectory
    ◯ MoveFile
    ◯ ImageGeneration
```
**Action:** Select PythonCode, ExecuteShellCommand, ReadTextFile, WriteFile

**Narration:**
> "AgentScope provides a rich set of built-in tools. For our research agent, we'll enable code execution for repository analysis, shell commands for git operations, and file I/O for generating reports."

---

**Step 6: MCP Server Configuration**
```
🔗 MCP Server Configuration
  Configure multiple MCP servers if you want to use external tools

🔗 Get existing MCP Servers from Himarket:
? Choose MCP Server from Current Himarket (http://localhost:5174)
  ❯ ◉ deepwiki-product
    ◯ weather-api-demo
    ◯ database-query-tool
```
**Action:** Select deepwiki-product

```
🔗 Get existing MCP Servers from Higress:
? Choose MCP Server from Current Higress (http://localhost:8080)
  ❯ ◉ arxiv-search
    ◉ deepwiki
    ◯ higress-api
```
**Action:** Select arxiv-search and deepwiki

**Narration:**
> "This is powerful - Hgctl automatically queries both Himarket and Higress Console to show all available MCP servers. We select the DeepWiki and ArXiv tools we configured earlier. No need to manually input URLs or authentication details."

```
Add MCP Servers manually...
? MCP Server URL (or press Enter to finish):
> [press Enter]
```

**Narration:**
> "We could add additional MCP servers manually with custom headers for authentication, but our selected tools are sufficient."

---

**Step 7: Deployment Settings**
```
🌐 Deployment Settings
  Network configuration for the agent
? Deployment port: (8090)
> 8090

? Host binding: (0.0.0.0)
> 0.0.0.0

  How the agent responds to user input
? Enable streaming responses? (Y/n)
> Y
```

**Narration:**
> "We configure the deployment port and enable streaming responses for real-time output. The agent will run as an HTTP service compatible with both OpenAI and Agent-to-Agent protocols."

---

**Step 8: Configuration Summary**
```
📊 Agent Configuration Summary:
  📝 Name: deepresearch
  🏢 Provider: AnthropicChat
  🤖 Model: claude-3-5-sonnet-20241022
  🔧 Tools: 4 selected
  🌐 Port: 8090
  📍 Host: 0.0.0.0
  ✨ Streaming: true
  🔗 MCP Servers: 3
    1. deepwiki-product - http://10.96.0.1:80/mcp-servers/deepwiki-product
    2. arxiv-search - http://10.96.0.1:80/mcp-servers/arxiv-search
    3. deepwiki - http://10.96.0.1:80/mcp-servers/deepwiki

✅ Configuration complete!
```

**Narration:**
> "Here's our complete configuration summary. The agent has all the tools needed for deep repository research."

---

### Scene 4.3: Agent Code Generation & Deployment (45 seconds)

**Continued Output:**
```
🔄 Generating agent code from AgentScope template...
  ✅ Rendered agent.py from template
  📁 Saved to ~/.hgctl/agents/deepresearch/agent.py

🐍 Checking Python environment...
  ✅ Python 3.11.5 found

📦 Checking dependencies...
  ℹ️  Virtual environment not found
  🔨 Creating virtual environment at ~/.hgctl/.venv...
  ✅ Virtual environment created
  📥 Installing agentscope[runtime]==0.1.0...
  ✅ Dependencies installed

🚀 Deploying agent...
  Starting agent service on 0.0.0.0:8090...

  ✅ Agent 'deepresearch' is now running!

📍 Access URLs:
  - OpenAI Compatible: http://127.0.0.1:8090/v1/chat/completions
  - Agent-to-Agent: http://127.0.0.1:8090/a2a/invoke

💡 Management Commands:
  - Stop: pkill -f "deepresearch/agent.py"
  - Logs: tail -f ~/.hgctl/agents/deepresearch/agent.log
  - Config: cat ~/.hgctl/agents/deepresearch/config.json
```

**Narration:**
> "Hgctl generated the complete agent code from the AgentScope template, created an isolated virtual environment, installed dependencies, and deployed the agent as a background service. The agent is now listening on port 8090 with OpenAI-compatible and A2A endpoints."

**Action:** Quick verification
```bash
curl http://127.0.0.1:8090/health
```

**Expected Output:**
```json
{
  "status": "healthy",
  "agent": "deepresearch",
  "version": "1.0.0",
  "uptime": "5s"
}
```

---

## Act 5: Publishing Agent to Higress & Himarket (2-3 min)

### Scene 5.1: Test Agent Locally (30 seconds)
**Narration:**
> "Before publishing to production, let's test the agent locally with a sample request."

**Command:**
```bash
curl -X POST http://127.0.0.1:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {
        "role": "user",
        "content": "Analyze the facebook/react repository and give me a brief summary"
      }
    ],
    "stream": false
  }' | jq .
```

**Expected Output (streaming text, show key parts):**
```json
{
  "id": "deepresearch-20250120-001",
  "object": "chat.completion",
  "model": "claude-3-5-sonnet-20241022",
  "choices": [
    {
      "message": {
        "role": "assistant",
        "content": "I'll analyze the React repository using the DeepWiki tool...\n\n[Tool call: deepwiki.read_wiki_structure('facebook/react')]\n\n# React Repository Analysis\n\n## Executive Summary\nReact is Facebook's declarative JavaScript library for building user interfaces...\n\n## Tech Stack\n- Core: JavaScript, Flow\n- Build: Rollup, Webpack\n- Testing: Jest, React Testing Library\n...\n"
      }
    }
  ]
}
```

**Narration:**
> "Excellent! The agent successfully used the DeepWiki MCP tool to analyze React's repository and generated a structured report. Now let's publish this to Higress as a production API."

---

### Scene 5.2: Publish to Higress as Model API (1 min)
**Command:**
```bash
hgctl agent add deepresearch http://127.0.0.1:8090 \
  --type model \
  --scope project \
  --description "DeepResearch Agent - GitHub repository analysis and research synthesis"
```

**Expected Output:**
```
🔍 Validating agent endpoint...
  ✅ Agent is healthy and responding

📋 Agent API Configuration:
  Name: deepresearch
  Type: model (OpenAI-compatible)
  Scope: project
  Source URL: http://127.0.0.1:8090

🚀 Publishing to Higress Console...

  📡 Creating AI Provider Service...
  POST /v1/ai/providers
  {
    "name": "deepresearch",
    "type": "custom",
    "endpoints": [{
      "url": "http://127.0.0.1:8090",
      "weight": 100
    }]
  }
  ✅ Provider service created

  🛣️  Creating AI Route...
  POST /v1/ai/routes
  {
    "name": "deepresearch",
    "path": "/v1/ai/deepresearch",
    "provider": "deepresearch",
    "models": ["claude-3-5-sonnet-20241022"]
  }
  ✅ AI route created

✅ Agent 'deepresearch' is published to Higress successfully!

📍 Access via Higress Gateway:
  http://10.96.0.1:80/v1/ai/deepresearch/chat/completions

🔧 Higress Console:
  http://localhost:8080/ai-routes/deepresearch
```

**Narration:**
> "The agent is now published to Higress as an AI Provider Service with a corresponding AI Route. Higress handles routing, load balancing, authentication, rate limiting, and observability. Requests go through the gateway at /v1/ai/deepresearch."

**Visual:** Quick browser view of Higress Console showing the new AI route

---

### Scene 5.3: Publish to Himarket as API Product (1-1.5 min)
**Command:**
```bash
hgctl agent add deepresearch-api http://127.0.0.1:8090 \
  --type model \
  --scope project \
  --as-product \
  --description "DeepResearch Agent API - Enterprise GitHub repository research and analysis service"
```

**Expected Output:**
```
🔍 Validating agent endpoint...
  ✅ Agent is healthy

🚀 Publishing to Higress Console...
  ✅ AI Provider service created: deepresearch-api
  ✅ AI Route created: /v1/ai/deepresearch-api

📦 Publishing to Himarket as API Product...

  🏷️  Product Type: MODEL_API
  📝 Product Details:
    - Name: deepresearch-api
    - Category: AI Agent Services
    - Type: MODEL_API
    - Description: Enterprise GitHub repository research and analysis service

  Creating product in Himarket...
  POST /api/v1/products
  {
    "name": "deepresearch-api",
    "type": "MODEL_API",
    "description": "DeepResearch Agent API - Enterprise GitHub repository research and analysis service",
    "aiConfig": {
      "provider": "deepresearch-api",
      "models": ["claude-3-5-sonnet-20241022"],
      "endpoint": "/v1/ai/deepresearch-api/chat/completions"
    }
  }
  ✅ Product created (ID: prod-agent-deepresearch-20250120)

  🔗 Linking to Higress Gateway...
  POST /api/v1/products/prod-agent-deepresearch-20250120/ref
  {
    "gatewayId": "higress-gateway-main",
    "gatewayName": "Higress Production Gateway"
  }
  ✅ Gateway reference established

✅ Agent 'deepresearch-api' is published to Himarket successfully!

📍 URLs:
  - Himarket Portal: http://localhost:5174/products/model-api/deepresearch-api
  - API Endpoint: http://10.96.0.1:80/v1/ai/deepresearch-api/chat/completions
  - API Docs: http://localhost:5174/docs/deepresearch-api
```

**Narration:**
> "Perfect! The DeepResearch Agent is now available as an API Product in Himarket. Let me show you what this looks like in the developer portal."

**Visual:** Switch to browser showing Himarket

**Browser Actions (screen record with narration):**

1. **Product Listing Page**
   - Navigate to http://localhost:5174/products
   - Show the "deepresearch-api" product in the Model API category
   - Narration: "Developers can browse available AI agent APIs by category"

2. **Product Detail Page**
   - Click on deepresearch-api product
   - Show product overview, description, pricing tier, and capabilities
   - Narration: "The product page shows comprehensive documentation auto-generated from the agent's configuration"

3. **API Documentation**
   - Click on "API Reference" tab
   - Show OpenAPI-style documentation with example requests/responses
   - Narration: "Full API documentation with request schemas and examples"

4. **Try It Out**
   - Click "Try API" button
   - Enter sample request:
     ```json
     {
       "messages": [
         {
           "role": "user",
           "content": "Analyze the vuejs/core repository"
         }
       ]
     }
     ```
   - Click "Send Request"
   - Show streaming response
   - Narration: "Developers can test the API directly from the portal before integrating"

5. **Usage Metrics**
   - Click on "Analytics" tab
   - Show request counts, latency graphs, error rates
   - Narration: "Himarket provides built-in observability - request metrics, latency tracking, and error monitoring"

---

## Act 6: End-to-End Demo & Conclusion (2 min)

### Scene 6.1: Complete Workflow Demo (1 min)
**Narration:**
> "Let's demonstrate the complete workflow with a real research request through the production API."

**Command:**
```bash
curl -X POST http://10.96.0.1:80/v1/ai/deepresearch-api/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_KEY}" \
  -d '{
    "messages": [
      {
        "role": "user",
        "content": "Generate a comprehensive research report for the anthropics/anthropic-sdk-python repository, including related academic papers"
      }
    ],
    "stream": true
  }'
```

**Expected Output (show streaming response, key highlights):**
```
data: {"choices":[{"delta":{"content":"I'll analyze the Anthropic SDK repository...\n\n"}}]}

data: {"choices":[{"delta":{"content":"[Using DeepWiki to analyze repository structure...]\n"}}]}

data: {"choices":[{"delta":{"content":"[Using ArXiv search for related papers...]\n\n"}}]}

data: {"choices":[{"delta":{"content":"# Anthropic Python SDK - Research Report\n\n"}}]}

data: {"choices":[{"delta":{"content":"## Executive Summary\n"}}]}
data: {"choices":[{"delta":{"content":"The anthropic-sdk-python is the official Python library for Anthropic's Claude API...\n\n"}}]}

data: {"choices":[{"delta":{"content":"## Architecture\n"}}]}
data: {"choices":[{"delta":{"content":"- **Type Safety**: Uses Pydantic for request/response validation\n"}}]}
data: {"choices":[{"delta":{"content":"- **Async Support**: Native asyncio implementation...\n\n"}}]}

data: {"choices":[{"delta":{"content":"## Related Academic Research\n"}}]}
data: {"choices":[{"delta":{"content":"1. **Constitutional AI**: [arXiv:2212.08073]\n"}}]}
data: {"choices":[{"delta":{"content":"2. **Tool Use and Function Calling**: [arXiv:2310.03744]\n\n"}}]}

data: {"choices":[{"delta":{"content":"## Development Health\n"}}]}
data: {"choices":[{"delta":{"content":"- Active maintenance with 450+ commits\n"}}]}
data: {"choices":[{"delta":{"content":"- Comprehensive test coverage (92%)\n\n"}}]}

data: [DONE]
```

**Narration:**
> "The agent is working through the complete research pipeline: analyzing the repository structure with DeepWiki, searching for related academic papers on ArXiv, and synthesizing everything into a comprehensive markdown report. All requests are routed through Higress, which handles authentication, rate limiting, and metrics collection."

---

### Scene 6.2: Architecture Summary (30 seconds)
**Visual:** Show architecture diagram with data flow highlighted

**Narration:**
> "Let's review what we've built:
>
> 1. **Tool Layer**: MCP servers (DeepWiki, ArXiv) provide standardized tool interfaces - both HTTP direct and OpenAPI-converted types
> 2. **Agent Layer**: DeepResearch agent built with AgentScope, using Claude as the reasoning engine, integrating multiple MCP tools
> 3. **Gateway Layer**: Higress provides production-grade routing, authentication, rate limiting, and observability
> 4. **Marketplace Layer**: Himarket exposes the agent as an API Product with documentation, testing playground, and usage analytics
>
> All of this was achieved with simple hgctl commands - no manual infrastructure configuration, no custom API wrappers, no complex deployment scripts."

---

### Scene 6.3: Competitive Advantages & Conclusion (30 seconds)
**Visual:** Side-by-side comparison or bullet points

**Narration:**
> "Hgctl Agent Module delivers on all competition requirements:
>
> ✅ **Solves Integration Hell**: MCP protocol standardizes tool integration - 2 commands to add any API (HTTP direct or OpenAPI conversion)
> ✅ **Low-Code Experience**: From zero to deployed agent in 5 minutes with interactive wizards
> ✅ **Production-Grade**: Higress provides enterprise routing, security, and governance out of the box
> ✅ **Open Ecosystem**: Published agents become discoverable, reusable API products in Himarket
> ✅ **Framework Agnostic**: Currently supports AgentScope, designed to extend to LangGraph, CrewAI, and others
>
> This isn't just a prototype tool - it's a complete platform for building, deploying, and managing production AI Agents at scale.
>
> Thank you for watching. The future of Agent development is here."

**Final Visual:**
- Terminal showing all running components
- Himarket dashboard with multiple API products
- Fade to hgctl logo

---

## Technical Notes for Recording

### Pre-Recording Setup Checklist
- [ ] Higress installed and running (localhost:8080)
- [ ] Himarket running (localhost:5174)
- [ ] Higress Gateway accessible (verify cluster IP)
- [ ] Claude Code installed and configured
- [ ] ~/.hgctl.json properly configured
- [ ] Environment variables set (ANTHROPIC_API_KEY, etc.)
- [ ] Sample OpenAPI spec file (arxiv-api-spec.yaml) prepared
- [ ] Terminal color scheme optimized for recording
- [ ] Browser tabs pre-opened to Higress Console and Himarket
- [ ] Test all commands in advance to verify output
- [ ] Prepare backup recordings for each section

### Recording Tips
1. **Terminal Settings**:
   - Use large font (16-18pt) for readability
   - Enable syntax highlighting
   - Use clear color scheme (light background recommended)
   - Set terminal size to 120x30 for consistent layout

2. **Command Execution**:
   - Type commands slowly or use pre-written scripts
   - Pause 2-3 seconds after each command output for viewer comprehension
   - Use `| jq` for JSON formatting where applicable
   - Clear terminal between major sections for visual clarity

3. **Browser Recording**:
   - Use browser zoom at 125-150% for UI elements
   - Highlight important elements with cursor movement
   - Scroll slowly through long content
   - Pause on key information

4. **Narration**:
   - Record voiceover separately for better quality
   - Speak clearly and at moderate pace
   - Emphasize key technical terms
   - Allow pauses for viewers to read terminal output

5. **Editing**:
   - Speed up long-running commands (installation, generation) to 2-3x
   - Use picture-in-picture for browser views while showing terminal
   - Add on-screen annotations for key concepts
   - Include timestamps for major sections

### Fallback Plans
- If live demo fails, have pre-recorded terminal sessions
- Prepare static screenshots for critical UI states
- Have alternative examples ready (different repositories for analysis)
- Keep simplified version of demo for time constraints

---

## Appendix: Key Commands Reference

```bash
# Configuration
cat ~/.hgctl.json
export HGCTL_AGENT_CORE=claude

# Start interactive window
hgctl agent

# Add HTTP MCP Server
hgctl mcp add <name> <url> --type http --transport streamable

# Add OpenAPI MCP Server
hgctl mcp add <name> <spec-file> --type openapi

# Add MCP Server as Product
hgctl mcp add <name> <url> --type http --as-product

# Create new agent
hgctl agent new

# Publish agent to Higress
hgctl agent add <name> <url> --type model --scope project

# Publish agent to Himarket
hgctl agent add <name> <url> --type model --as-product

# Test agent locally
curl -X POST http://127.0.0.1:8090/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"test"}]}'

# Test through Higress
curl -X POST http://10.96.0.1:80/v1/ai/<agent-name>/chat/completions \
  -H "Authorization: Bearer ${API_KEY}" \
  -d '{"messages":[{"role":"user","content":"test"}]}'
```
