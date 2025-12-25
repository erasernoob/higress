# Hgctl Agent 模块整体设计文档

## 一、背景

为了响应赛题指出的目前 Agent 构建的集成地狱：存量 API 的转化，以及工具与 Agent 的集成等等问
题，在此提出本方案。本方案基于 Higress 的现有的命令行管理工具 Hgctl，扩展并开发了全新的
Agent 模块。该模块通过引入 AgentCore 概念，将优秀的命令行 AI 工具（如 Claude Code、QoderCli
等）作为交互式 Agent 的底层引擎，实现了从想法到 Agent 的极速落地方案。

核心设计思想在于三个关键环节：

```
通过 MCP（Model Context Protocol）标准化工具接入流程，解决 Agent 与外部 API 的集成地狱问
题。
基于 AgentScope 等成熟框架提供模板化的 Agent 代码生成能力，降低开发⻔槛。
与 Higress Console 和 Himarket 深度集成，实现 Agent API 的一键发布与动态管理，确保生产级
的可观测性和治理能力。
```
## 二、 AgentCore 与 Hgctl 的深度集成

## 2.1 灵活配置体系

Hgctl Agent 模块采用了符合命令行工具最佳实践的多层配置机制。配置信息可以通过三种方式提供：
JSON 配置文件（位于用戶目录 ~/.hgctl.json）、环境变量（支持自动大写和下划线转换），以及命令行
参数。

配置项涵盖了 AgentCore 的选择（claude 或 qodercli）、Higress Console 的认证信息（URL、用戶
名、密码）、Higress Gateway 的访问地址，以及 Himarket 管理平台的连接配置。 这种分层设计保证了
敏感信息的安全存储，并提供了灵活的运行时覆盖能力，非常适合在不同环境（开发、测试、生产）间
快速切换。

Agent 模块参考配置文件：


```
{
"hgctl-agent-core": "claude",
"higress-gateway-url": "http://localhost:80",
"higress-console-url": "http://localhost:8080",
"higress-console-user": "admin",
"higress-console-password": "123",
"higress-gateway-url": "http://127.0.0.1:80",
"himarket-admin-url": "http://localhost:5174",
"himarket-admin-user": "admin",
"himarket-admin-password": "123"
}
```
### 2.2 环境初始化

首次使用 Hgctl Agent 时，系统会自动检测并安装所需的运行环境。初始化流程分为三个阶段：

1. 检查 Node.js 环境（要求版本不低于 18 ），若未安装则根据操作系统类型提供自动安装或手动安装
    指引。
2. 验证选定的 AgentCore 是否可用，不可用时通过 npm 自动安装。
3. 初始化本地工作目录，将内置的 Agent 模板、命令定义等资源文件释放到 ~/.hgctl 和
    ~/.{agentcore} 目录下。

在初始化过程中会自动配置一个预置的 Higress-API MCP Server ，该 MCP Server 封装了 Higress
Console 的管理 API，使得 Agent 能够通过自然语言指令直接操作 Higress 的路由、服务、MCP
Server 等资源。简化了 Agent 与基础设施的交互，是实现低代码体验的关键一环。

### 2.3 交互式 Agent 窗口

通过执行 hgctl agent 命令，用戶可以启动基于 AgentCore 的交互式 Agent 窗口。该窗口提供了完整
的对话式开发体验，用戶可以用自然语言描述需求，Agent 会调用预先注册的 MCP Server 完成相应操
作。交互窗口的底层实现通过 Go 的 exec 包调用 AgentCore 的二进制文件，确保了进程间的标准输入
输出流畅传递，实现了用戶与 Agent 的自然对话。

## 三、 MCP Server 管理

MCP（Model Context Protocol）是 Agent 与外部工具交互的标准化协议。Hgctl Agent 模块提供了两种
类型的 MCP Server 管理能力，分别对应不同的集成场景。


### 3.1 HTTP 类型：直接代理现有服务

##### 对于已经实现了 MCP 协议的服务端点，可以通过 HTTP 类型快速接入。执行命令：

```
hgctl mcp add http-mcp http://localhost:8080/mcp --type http --transport streamable
```
此命令会将指定的 HTTP 端点注册到 AgentCore 的配置中，并可选地发布到 Higress Console 。发布
过程包括创建服务源（Service Source）和 MCP Server 资源，使得该工具可以被集群内其他 Agent 共
享使用。

参数 --no-publish 可以控制是否发布到 Higress，若仅希望在本地交互窗口使用该工具而不影响集群
配置，可以加上此标志。 并且HTTP 类型 MCP Server 支持添加自定义 HTTP 请求头，以及对应环境变
量，这对于需要认证的 API 非常有用。

### 3.2 OpenAPI 类型：自动转换 REST API

对于传统的 REST API，Hgctl 提供了基于 OpenAPI 规范的自动转换能力。用戶只需提供 OpenAPI 文
档（YAML 或 JSON 格式），系统会自动将其转换为 MCP Server 配置。执行命令：

```
hgctl mcp add swagger-mcp ./path/to/openapi.yaml --type openapi
```
转换流程分为三个步骤：首先使用 Higress-Group 提供的 openapi-to-mcpserver 工具解析 OpenAPI 文
档，提取出所有的 API 端点、参数定义、数据模型；然后将解析结果转换为 MCP 配置格式；最后调用
Higress Console API 创建对应的 MCP Server 和 OpenAPI Tool 配置。

OpenAPI 类型的 MCP Server 必须发布到 Higress，因为其功能依赖 Higress 的 MCP 插件在运行时动
态调用后端 API。发布完成后，系统会自动获取 Higress Gateway 的集群内访问地址，并将生成的
MCP Server 端点（形如 [http://{gateway-ip}/mcp-servers/{name}](http://{gateway-ip}/mcp-servers/{name}) ）注册到 AgentCore，从而实现了
本地交互窗口与集群资源的无缝衔接。

### 3.3 与 Himarket 的集成： API ⻔戶化管理

Hgctl 支持一键将 MCP Server 作为 API Product 发布到 Himarket 平台。通过添加 --as-product 参
数：

```
hgctl mcp add my-mcp http://api.example.com/mcp --as-product
```
系统会在 Himarket 中创建一个类型为 MCP_SERVER 的 API Product，并将其关联到指定的 Higress
网关实例。开发者可以通过 Himarket ⻔戶网站浏览可用工具、申请访问权限、查看调用统计等。


## 四、 Agent 创建与部署

### 4.1 设计理念

Hgctl Agent 模块的创建流程基于"模板 + 配置"的设计模式。系统内置了基于 AgentScope 框架的
Python 模板，该模板已经实现了完整的 Agent 生命周期管理，包括状态持久化、会话历史、工具注
册、 **MCP** 客戶端集成、流式输出等核心功能。开发者无需关心这些底层细节，只需要通过交互式问答
提供业务配置信息，系统就能生成可运行的 Agent 代码。

同时支持通过插件机制扩展对其他 Agent 框架（如 LangGraph、CrewAI 等）的支持，只需添加新的模
板文件和对应的渲染逻辑即可。

### 4.2 交互式 Agent 创建流程

执行 hgctl agent new 命令会启动一个友好的交互式创建向导。用戶首先选择创建方式：从头开始逐步
配置（create step by step），或者从 AgentCore 中导入已有的 Agent 定义（import existing one）。

#### 4.2.1 从头创建

##### 选择从头创建后，系统会依次询问以下配置项：

Agent 基本信息包括 Agent 名称和应用描述，名称将作为目录名和服务标识。

系统 Prompt 配置支持三种输入方式：

```
直接文本输入
指定本地 Markdown 文件路径
通过 LLM 自动生成。
```
在第三种方式下，用戶只需用自然语言描述 Agent 的预期行为（如"帮我编写单元测试"），系统会调用

预置的 (^) gen-agent 命令 ，利用 AgentCore 的 LLM 能力生成结构化的 Prompt 文本。充分利用了 LLM
的语义理解和内容生成能力，降低了 Prompt 工程的⻔槛。
工具选择环节会展示 AgentScope 框架支持的所有内置工具（如代码执行、文件操作、文生图、图生文
等），用戶可以通过多选方式勾选需要的工具。
模型配置允许指定使用的 LLM 模型名称（默认 qwen-flash）和 API Key 的环境变量名。 同时支持通过
环境变量 AGENT_CHAT_MODEL 覆盖配置。
部署参数包括:
监听端口


##### 主机绑定地址

##### 是否启用流式响应

##### 思维链模式

```
MCP Server 配置
```
MCP Server 集成具体逻辑：

```
系统会先自动查询当前 Higress Console 中已注册的所有 MCP Server，用戶可以从列表中直接勾
选需要的工具，无需手动输入 URL 和配置。Agent 创建时就能"看到"集群中可用的所有工具资源。
可以手动添加额外的 MCP Server（如第三方服务或内部测试环境的端点），每个 MCP Server 可以
配置自定义 HTTP 请求头用于认证。
```
所有配置收集完成后，系统会显示一份配置摘要供用戶确认，然后基于 AgentScope 模板生成 Agent 代
码，保存到 ~/.hgctl/agents/{agent-name}/agent.py 路径下。

#### 4.2.2 从 AgentCore 导入

对于在 AgentCore 交互窗口中已经创建并调试过的 Agent，可以通过导入模式快速迁移。
系统会扫描 AgentCore 的 agents 目录（如 ~/.claude/agents/），列出所有可用的 Agent 定义文件
（Markdown 格式），用戶选择后即可将其 Prompt 内容复制到 Hgctl 的管理目录下。开发者可以在
AgentCore 中快速原型验证，在 Hgctl 中实现生产化部署。

#### 4.3 自动化部署与环境管理

Agent 代码生成后，系统会自动启动部署流程。部署过程首先检查 Python 运行环境（要求 3.10 及以上
版本），然后验证必要的依赖包（agentscope 和 agentscope-runtime）是否已安装。

若依赖缺失，系统会创建一个独立的 Python 虚拟环境（位于 ~/.hgctl/.venv ），并在其中安装所需依
赖。虚拟环境只需创建一次，后续部署会复用已有环境，加快了启动速度。

部署执行时，系统会在虚拟环境中运行生成的 agent.py 文件。Agent 进程会以 detached 模式启动，
监听在本地配置的端口上，并输出访问 URL 和管理命令提示。此时 Agent 已经作为一个完整的 HTTP
服务运行。依托于 agentscope-runtime 强大部署能力，该独立运行的服务支持 A2A（Agent-to-Agent）
协议以及 OpenAI 兼容格式的请求。

### 4.4 发布到 Higress 与 Himarket

Agent 本地部署成功后，下一步是将其发布到 Higress 网关，使其成为可被外部访问和管理的 API 服
务。执行命令：

```
hgctl agent add my-agent http://127.0.0.1:8090 --type model --scope project
```

其中 --type 参数支持三种 API 类型：

```
a2a 用于 Agent-to-Agent 协议
model 用于 OpenAI 兼容的模型 API
restful 用于标准 REST API。
```
相对应的，不同类型会在 Higress 中创建不同的路由和服务配置

对于 model 类型，系统会在 Higress Console 中创建一个 Provider Service 和对应的 AI Route，使得
Agent 可以通过标准的 /v1/responses 端点被访问。对于 restful 类型，则会创建普通的服务源和路由配
置。

参数 --scope 控制配置的作用范围，project 表示项目级别（默认值），global 或 user 表示全局或用戶
级别。

发布过程同样支持 --no-publish 和 --as-product 参数。 前者用于仅在本地 AgentCore 中注册该 Agent
而不影响 Higress 配置，后者用于将 Agent API 同步发布到 Himarket 作为 API Product。发布到
Himarket 时，系统会创建对应类型的 Product（如 AGENT_API 或 MODEL_API），并建立与 Higress
网关实例的引用关系，使得该 Agent 可以在 Himarket ⻔戶中被发现、订阅和监控。

## 五、补充机制

### 5.1 预置 subAgent 与命令

Hgctl 内置了预定义的 Agent 和命令，只要使用 hgctl agent 启动 TUI 主 Agent 就会自带
openapi-generator Agent，可以根据 HTTP 端点自动生成 OpenAPI 规范文档；

(^) gen-agent 命令 则用于生成高质量的 Agent Prompt。这些预置资源在系统初始化时会被自动释放到用
戶目录和 AgentCore 目录下，用戶可以直接在交互窗口中调用。

### 5.2 凭证自动发现

Hgctl 实现了智能的凭证发现机制：若当前 Higress 是通过 hgctl install 安装的，系统会自动从
Kubernetes Secret 中读取 Console 的用戶名和密码；若凭证读取失败或用戶希望连接到其他 Higress
实例，系统会通过交互式提示引导用戶输入。

## 六、补充说明

Hgctl Agent 模块通过将命令行工具、AI Agent 框架、MCP 标准和 Higress 网关深度整合，将 Agent 开
发⻔槛降至最低；与 Higress 和 Himarket 的深度集成使得 Agent 天然具备路由、认证、限流、监控等


##### 治理能力。较为完整的提供了一个赛题所期望的目标方案，但仍然需要迭代优化:

```
支持更多 Agent 框架（LangGraph、CrewAI 等）的模板；
实现