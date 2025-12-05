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
		Use:   "create agent [name]",
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

			fmt.Printf("Agent '%s' created successfully!\n", name)
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
