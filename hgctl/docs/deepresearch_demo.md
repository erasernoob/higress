## DeepResearch Screen Play

### Predefined

- Show configuration details
    - 灵活的配置信息（三种方式：环境变量、JSON 配置文件、命令行参数）
    - 多用户配置支持（Higress 管理员、Himarket 开发者、Himarket 管理者)
    - 适合在不同的环境（开发、测试、生产）快速切换
- Environment initial
    - AgenticCore 自动检测（优先级、qodercli > cc、支持显式配置)
    - 支持提示用户，并自动安装 
    - 检查运行环境并自动安装（Nodejs... Python ）
    - 初始化本地工作目录 
    - 内置 subagent、slash commands、mcp server （Deepwiki、Higress-api...）提前注入
- Add needed MCP（MCP Server Management： Higress Himarket）
    - HTTP 类型
    - Openapi 类型
> 这里需要添加一个 MCP Server 既可以用于展示功能，又能在最后创建 Agent 的时候复用呢？
> 1. deepwiki
> 2. openapi 类型的？

- Create DeepResearch Agent use claude code
- after create the agent use launch the agentic core to test and improve the agent's functionality
- After test and improvement, deploy the agent API to higress and himarket
- Then test it on himarket's web UI
