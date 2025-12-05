package agent

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/alibaba/higress/hgctl/pkg/manifests"
	"github.com/alibaba/higress/hgctl/pkg/util"
	"github.com/spf13/cobra"
)

var ASAvailiableTools = []string{
	"execute_python_code",
	"execute_shell_command",
	"view_text_file",
	"write_text_file",
	"insert_text_file",
	"dashscope_text_to_image",
	"dashscope_text_to_audio",
	"dashscope_image_to_text",
	"openai_text_to_image",
	"openai_text_to_audio",
	"openai_edit_image",
	"openai_create_image_variation",
	"openai_image_to_text",
	"openai_audio_to_text",
}

type MCPServerConfig struct {
	Name      string            // MCP Client Name
	URL       string            // MCP Server URL
	Transport string            // transport `streamable_http` or `see` or `stdio`
	Headers   map[string]string // HTTP Headers
}

type AgentConfig struct {
	AppName         string   //  "app"
	AppDescription  string   //  "A helpful assistant and useful agent"
	AgentName       string   //  "Friday"
	AvailableTools  []string //   availiable tools (built-in agentscope)
	SysPrompt       string   //  "You are a helpful assistant"
	ChatModel       string   //  "qwen-max"
	APIKeyEnvVar    string   //  DASHCOPE_API_KEY
	DeploymentPort  int      //  8090
	HostBinding     string   //  0.0.0.0
	EnableStreaming bool     //  true
	EnableThinking  bool     //  true
	MCPServers      []MCPServerConfig
}

func createAgentCmd() *cobra.Command {
	var createAgentCmd = &cobra.Command{
		Use:   "new agent [name]",
		Short: "Create a new agent",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]

			config, err := getAgentConfig(name)
			if err != nil {
				fmt.Printf("Error get Agent config: %v\n", err)
				os.Exit(1)
			}
			if err := generateAgent(config); err != nil {
				fmt.Printf("Error creating agent: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Agent '%s' created successfully! Running...\n", name)
			runAgent(name)

		},
	}
	return createAgentCmd

}

func generateAgent(config *AgentConfig) error {
	templateStr, err := get_agentscope_template()
	if err != nil {
		return fmt.Errorf("failed to read template: %v", err)
	}

	// sync with python
	funcMap := template.FuncMap{
		"boolToPython": func(b bool) string {
			if b {
				return "True"
			}
			return "False"
		},
	}

	tmpl, err := template.New("agent").Funcs(funcMap).Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	agentsDir := util.GetHomeHgctlDir() + "/agents"
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %v", err)
	}

	agentDir := filepath.Join(agentsDir, config.AgentName)
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		return fmt.Errorf("failed to create agent directory: %v", err)
	}

	outputFile := filepath.Join(agentDir, "agent.py")
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, config); err != nil {
		return fmt.Errorf("failed to render template: %v", err)
	}

	return nil
}

func get_agentscope_template() (string, error) {
	f := manifests.BuiltinOrDir("")
	const templatePath = "agent/template/agentscope.tmpl"
	data, err := fs.ReadFile(f, templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}

	return string(data), nil
}

type AgentEnvCheck struct {
	PythonVersion       string
	PythonExists        bool
	AgentscopeInstalled bool
	RequiredDeps        []string
	MissingDeps         []string
}

func checkAgentEnvironment(agentName string) (*AgentEnvCheck, error) {
	check := &AgentEnvCheck{
		RequiredDeps: []string{
			"agentscope",
			"agentscope-runtime==1.0.0",
		},
	}

	pyVenv, err := util.GetPythonVersion()
	if err != nil {
		fmt.Printf("Python environment not found, you need Python environment to deploy your agent\n")
		return nil, err
	}

	if util.CompareVersions(pyVenv, "3.10") == -1 {
		fmt.Printf("Current Python: %s need Python 3.10+", pyVenv)
		return nil, fmt.Errorf("unsupport python version")
	}

	return check, checkDeps()
}

func checkDeps() error {
	// Currently check agentscope and agentscopeRuntime
	cmd := exec.Command("python3", "-c", "import agentscope; print(agentscope.__version__)")
	_, err := cmd.Output()
	if err != nil {
		fmt.Printf("agentscope not installed: %s, installing...", err)
		cmd := exec.Command("pip", "install", "agentscope")
		if _, err = cmd.Output(); err != nil {
			return fmt.Errorf("failed to install agentscope: %e", err)
		}
	}

	cmd = exec.Command("python3", "-c", "import agentscope_runtime; print(agentscope_runtime.__version__)")
	_, err = cmd.Output()
	if err != nil {
		fmt.Printf("agentscope-runtime not installed: %w, installing...", err)
		cmd := exec.Command("pip", "install", "agentscope-runtime==1.0.0")
		if _, err = cmd.Output(); err != nil {
			return fmt.Errorf("failed to install agentscope-runtime: %e", err)
		}
	}

	return nil

}

func runAgent(agentName string) error {

	_, err := checkAgentEnvironment(agentName)
	if err != nil {
		return fmt.Errorf("environment check failed: %w", err)
	}

	agentDir := filepath.Join(util.GetHomeHgctlDir(), "agents")
	agentDir = filepath.Join(agentDir, agentName, "agent.py")

	if _, err := os.Stat(agentDir); os.IsNotExist(err) {
		return fmt.Errorf("agent file not found: %s", agentDir)
	}

	// if len(envCheck.MissingDeps) > 0 {
	// 	fmt.Println(color.YellowString("⚠️  Some dependencies are missing. Installing..."))
	// 	if err := installAgentDependencies(agentName); err != nil {
	// 		return fmt.Errorf("dependency installation failed: %w", err)
	// 	}
	// } else {
	// 	fmt.Println(color.GreenString("✅ All dependencies already installed!"))
	// }
	// fmt.Println()

	return startAgentProcess(agentDir, agentName)
}

func startAgentProcess(agentPath, agentName string) error {
	switch runtime.GOOS {
	case "windows":
		return runWindowsAgent(agentPath, agentName)
	case "darwin", "linux":
		return runUnixAgent(agentPath, agentName)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func runUnixAgent(agentPath, agentName string) error {
	// // 检查端口是否被占用
	// port, err := getAgentPort(agentName)
	// if err != nil {
	// 	return fmt.Errorf("failed to get agent port: %w", err)
	// }

	// if isPortInUse(port) {
	// 	color.Yellow("⚠️  Port %d is already in use", port)
	// 	return fmt.Errorf("port %d is already in use", port)
	// }

	// 启动Python agent
	cmd := exec.Command("python3", agentPath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	return cmd.Wait()
}

func runWindowsAgent(agentPath, agentName string) error {
	cmd := exec.Command("python3", agentPath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	return cmd.Wait()
}

func getAgentPort(agentName string) (int, error) {
	// 这里可以从配置文件中读取端口号
	// 默认返回8090
	return 8090, nil
}

func isPortInUse(port int) bool {
	// 检查端口是否被占用
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port))
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
