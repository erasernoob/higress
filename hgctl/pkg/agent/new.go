package agent

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"github.com/alibaba/higress/hgctl/pkg/manifests"
	"github.com/alibaba/higress/hgctl/pkg/util"
	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

const (
	ASRuntimeMainPyFile = "as_runtime_main.py"
	AgentRunMainPyFile  = "agentrun_main.py"
	ToolKitPyFile       = "toolkit.py"
	SConfigYAML         = "s.yaml"

	ARTemplate      = "agentrun.tmpl"
	ASTemplate      = "agentscope.tmpl"
	ToolKitTemplate = "toolkit.tmpl"
	SConfigTemplate = "agentrun_s.tmpl"
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
	MinPythonVersion = "3.12"

	DefaultServerLessAccessKey = "hgctl-credential"
)

type MCPServerConfig struct {
	Name      string            // MCP Client Name
	URL       string            // MCP Server URL
	Transport string            // transport `streamable_http` or `see` or `stdio`
	Headers   map[string]string // HTTP Headers
}

type ServerlessConfig struct {
	AccessKey    string
	ResourceName string
	Region       string
	AgentName    string
	AgentDesc    string
	Port         uint

	DiskSize uint
	Timeout  uint

	GlobalConfig HgctlAgentConfig
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

	Type          DeployType
	ServerlessCfg ServerlessConfig
}

func createAgentCmd() *cobra.Command {
	agentRun := false
	deployDirect := false

	var createAgentCmd = &cobra.Command{
		Use:   "new",
		Short: "Create a new agent or import one from core",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			config := &AgentConfig{
				Type: Local,
			}
			if agentRun {
				config.Type = AgentRun
				config.ServerlessCfg = ServerlessConfig{
					AccessKey: DefaultServerLessAccessKey,
					Port:      9000,
					DiskSize:  512,
					Timeout:   600,

					GlobalConfig: GlobalConfig,
				}
			}

			if err := getAgentConfig(config); err != nil {
				fmt.Printf("Error get Agent config: %v\n", err)
				os.Exit(1)
			}

			if err := createAgentTemplate(config); err != nil {
				fmt.Printf("Error creating agent: %v\n", err)
				os.Exit(1)
			}

			if deployDirect {
				handler := &DeployHandler{Name: config.AgentName}
				cmdutil.CheckErr(handler.Deploy())
			} else {
				fmt.Printf("🌟 agent %s created successfully, you can deploy it by using `hgctl agent deploy %s`\n", config.AgentName, config.AgentName)
			}

		},
	}

	createAgentCmd.PersistentFlags().BoolVar(&agentRun, "agent-run", false, "Use agentRun to deploy to Alibaba cloud, default is false")
	createAgentCmd.PersistentFlags().BoolVar(&deployDirect, "deploy", false, "After agent creation, deploy it directly")
	return createAgentCmd

}

func createAgentTemplate(config *AgentConfig) error {
	agentsDir := util.GetHomeHgctlDir() + "/agents"
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %v", err)
	}

	agentDir := filepath.Join(agentsDir, config.AgentName)
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		return fmt.Errorf("failed to create agent directory: %v", err)
	}

	switch config.Type {
	case Local:
		// parse agentscope file
		asMain := filepath.Join(agentDir, ASRuntimeMainPyFile)
		asTemplateStr, err := get_template(ASTemplate)
		if err != nil {
			return fmt.Errorf("failed to read agentscope template: %v", err)
		}
		if err := renderTemplateFile(asTemplateStr, asMain, config); err != nil {
			return fmt.Errorf("failed to render agentscope runtime's file: %s", err)
		}
	case AgentRun:
		// Details see: https://github.com/Serverless-Devs/agentrun-sdk-python

		// parse agentrun file
		arMain := filepath.Join(agentDir, AgentRunMainPyFile)
		arTemplateStr, err := get_template(ARTemplate)
		if err != nil {
			return fmt.Errorf("failed to read agentrun template: %v", err)
		}
		if err := renderTemplateFile(arTemplateStr, arMain, config); err != nil {
			return fmt.Errorf("failed to render agentscope runtime's file: %s", err)
		}

		// parse s.yaml
		s := filepath.Join(agentDir, SConfigYAML)
		STmplStr, err := get_template(SConfigTemplate)
		if err != nil {
			return fmt.Errorf("failed to read agentrun's serverless config file template: %v", err)
		}
		if err := renderTemplateFile(STmplStr, s, config.ServerlessCfg); err != nil {
			return fmt.Errorf("failed to render agentscope runtime's file: %s", err)
		}

		// write requirements
		fileContent := "agentrun-sdk[agentscope,server]>=0.0.3"
		targetFilePath := filepath.Join(agentDir, "requirements.txt")
		if err := util.WriteFileString(targetFilePath, fileContent, os.ModePerm); err != nil {
			return fmt.Errorf("failed to write requirements.txt file to target agent directory: %s", err)
		}
	}

	// parse toolkitPath
	toolkitPath := filepath.Join(agentDir, ToolKitPyFile)
	toolkitTmpl, err := get_template(ToolKitTemplate)
	if err != nil {
		return fmt.Errorf("failed to read toolkit template: %v", err)
	}
	if err := renderTemplateFile(toolkitTmpl, toolkitPath, config); err != nil {
		return fmt.Errorf("failed to render toolkit file: %s", err)
	}

	return nil
}

func renderTemplateFile(templateStr string, targetPath string, data interface{}) error {
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
	file, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to render template: %v", err)
	}

	return nil
}

func get_template(templatePath string) (string, error) {
	f := manifests.BuiltinOrDir("")
	templatePath = "agent/template/" + templatePath
	data, err := fs.ReadFile(f, templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}

	return string(data), nil
}
