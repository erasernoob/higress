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

const (
	MinPythonVersion = "3.10"
)

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
	SysPromptPath   string   //  "You are a helpful assistant"
	ChatModel       string   //  "qwen-max"
	APIKeyEnvVar    string   //  DASHCOPE_API_KEY
	DeploymentPort  int      //  8090
	HostBinding     string   //  0.0.0.0
	EnableStreaming bool     //  true
	EnableThinking  bool     //  true
	MCPServers      []MCPServerConfig
}

type AgentHandler struct {
	*AgentConfig

	PythonVenvPath string
	AgentDir       string
	AgentFile      string
}

func createAgentCmd() *cobra.Command {
	var createAgentCmd = &cobra.Command{
		Use:   "new agent",
		Short: "create a new agent or import one from core",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			config, err := getAgentConfig()
			if err != nil {
				fmt.Printf("Error get Agent config: %v\n", err)
				os.Exit(1)
			}

			if err := createAgentTemplate(config); err != nil {
				fmt.Printf("Error creating agent: %v\n", err)
				os.Exit(1)
			}

			agentDir := filepath.Join(util.GetHomeHgctlDir(), "agents")
			agentFile := filepath.Join(agentDir, config.AgentName, "agent.py")

			handler := &AgentHandler{
				AgentConfig:    config,
				PythonVenvPath: "",
				AgentDir:       agentDir,
				AgentFile:      agentFile,
			}

			fmt.Printf("Agent '%s' created successfully! Start to deploy it to local...\n", config.AgentName)
			if err := handler.runAgent(); err != nil {
				fmt.Printf("Error deploy agent: %v\n", err)
				os.Exit(1)
			}
		},
	}
	return createAgentCmd

}

func createAgentTemplate(config *AgentConfig) error {
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

func (h *AgentHandler) RunPythonCmd(showOutput bool, args ...string) error {
	pythonPath := filepath.Join(h.PythonVenvPath, "bin", "python3")
	cmd := exec.Command(pythonPath, args...)

	if showOutput {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (h *AgentHandler) checkAgentRequiredEnvironment() error {
	// requiredDeps := []string{
	// 	"agentscope",
	// 	"agentscope-runtime==1.0.0",
	// }

	pyVenv, err := util.GetPythonVersion()
	if err != nil {
		fmt.Printf("Python environment not found, you need Python environment to run your agent\n")
		return err
	}
	// TODO: not graceful way to find initial python path
	if path, err := exec.LookPath("python"); err == nil {
		pythonPathAbs, _ := filepath.Abs(path)
		venvBin := filepath.Dir(pythonPathAbs)
		venvRoot := filepath.Dir(venvBin)
		h.PythonVenvPath = venvRoot
	}

	if util.CompareVersions(pyVenv, MinPythonVersion) == -1 {
		fmt.Printf("Current Python: %s need Python 3.10+", pyVenv)
		return fmt.Errorf("unsupport python version")
	}

	if err := h.checkRequiredDeps(); err != nil {
		return err
	}

	return nil
}

// Currently only check agentscope and agentscope-runtime
func (h *AgentHandler) checkRequiredDeps() error {
	missingDeps := []string{}
	if err := h.RunPythonCmd(false, "-c", "import agentscope; print(agentscope.__version__)"); err != nil {
		fmt.Println("agentscope not installed, installing...")
		missingDeps = append(missingDeps, "agentscope")
	}

	if err := h.RunPythonCmd(false, "-c", "import agentscope_runtime; print(agentscope_runtime.__version__)"); err != nil {
		fmt.Println("agentscope-runtime not installed, installing...")
		missingDeps = append(missingDeps, "agentscope-runtime==1.0.0")
	}

	if len(missingDeps) != 0 {
		venvDir := filepath.Join(util.GetHomeHgctlDir(), ".venv")
		h.PythonVenvPath = venvDir

		if err := h.RunPythonCmd(true, "-m", "pip", "--version"); err != nil {
			fmt.Printf("Pip not installed, you need install pip to deploy your agent\n")
			return err
		}

		fmt.Println("This may takes a few minutes, you can install missing deps by yourself: ")
		for _, deps := range missingDeps {
			fmt.Println("- ", deps)
		}

		cmd := exec.Command("python3", "-m", "venv", venvDir)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("failed to create python virtual environment", string(output))
			return err
		}
		path := os.Getenv("PATH")
		newPath := venvDir + "/bin:" + path
		err = os.Setenv("PATH", newPath)
		if err != nil {
			fmt.Println("Failed to set PATH:", err)
			return err
		}
		err = os.Setenv("VIRTUAL_ENV", venvDir)
		if err != nil {
			fmt.Println("Failed to set VIRTUAL_ENV:", err)
			return err
		}
		for _, deps := range missingDeps {
			if err := h.RunPythonCmd(true, "-m", "pip", "install", deps); err != nil {
				fmt.Printf("failed to install missing deps: %s\n", deps)
				return err
			}

		}
		fmt.Println("Missing deps installed successfully, target python venv path: ", venvDir)
	}
	return nil
}

func (h *AgentHandler) runAgent() error {
	if err := h.checkAgentRequiredEnvironment(); err != nil {
		return fmt.Errorf("environment check failed: %w", err)
	}

	if _, err := os.Stat(h.AgentDir); os.IsNotExist(err) {
		return fmt.Errorf("agent source file not found: %s", h.AgentDir)
	}

	if err := h.startAgentProcess(); err != nil {
		return err
	}

	fmt.Printf(
		"🌟 You can deploy it to higress by using: hgctl agent add %s %s\n",
		h.AgentName, fmt.Sprintf("http://%s:%d", h.HostBinding, h.DeploymentPort))

	return nil

}

func (h *AgentHandler) startAgentProcess() error {
	switch runtime.GOOS {
	case "windows":
		return h.runWindowsAgent()
	case "darwin", "linux":
		return h.runUnixAgent()
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func (h *AgentHandler) runUnixAgent() error {
	fmt.Println(h.PythonVenvPath)
	if err := h.RunPythonCmd(true, h.AgentFile); err != nil {
		fmt.Println("failed to start agent, exiting...")
		return err
	}
	return nil
}

func (h *AgentHandler) runWindowsAgent() error {
	if err := h.RunPythonCmd(true, h.AgentFile); err != nil {
		fmt.Println("failed to start agent, exiting...")
		return err
	}
	return nil
}

func isPortInUse(port int) bool {
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port))
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
