// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package agent

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alibaba/higress/hgctl/pkg/manifests"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

type AgenticCore struct{}

func NewAgenticCore() *AgenticCore {
	core := &AgenticCore{}
	core.Setup()
	return core
}

func (c *AgenticCore) run(args ...string) error {
	cmd := exec.Command(AgentBinaryName, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()

}

// setup additional prequisite environment and plugins manifest to user's profile
// e.g. ../manifest/agent
func (c *AgenticCore) Setup() {
	embedFS := manifests.BuiltinOrDir("")
	homeDir, _ := os.UserHomeDir()
	if err := manifests.ExtractEmbedFiles(embedFS, "agent", filepath.Join(homeDir, ".hgctl")); err != nil {
		fmt.Println(err)
		fmt.Println("failed to init plugins for claude code")
		os.Exit(1)
	}

	if err := c.addHigressAPIMCP(); err != nil {
		fmt.Println("failed to init higress-api mcp server (you may need to add it manually): ", err)
		return
	}
}

func (c *AgenticCore) addHigressAPIMCP() error {
	arg := &HigressConsoleAuthArg{
		baseURL:    "",
		hgUser:     "",
		hgPassword: "",
	}
	fmt.Println("Automatically add Higress-api MCP server...")
	gatewayPrompt := promptui.Prompt{
		Label:   "Enter higress gateway URL",
		Default: "http://127.0.0.1:80",
	}
	gateway, err := gatewayPrompt.Run()
	if err != nil {
		fmt.Println("failed to run gateway prompt: ", err)
	}

	arg.baseURL = gateway

	if err := tryToGetLocalCredential(arg); err != nil {
		fmt.Println(err)
		// fallback: interact with user to provide password & username
		color.Red("failed to get higress-console credential automatically. Let's fix it manually")
		userPrompt := promptui.Prompt{
			Label:   "Enter higress console username",
			Default: "",
		}
		username, err := userPrompt.Run()
		if err != nil {
			return fmt.Errorf("aborted: %v", err)
		}
		pwdPrompt := promptui.Prompt{
			Label:   "Enter higress console password",
			Default: "",
		}
		pwd, err := pwdPrompt.Run()
		if err != nil {
			return fmt.Errorf("aborted: %v", err)
		}
		arg.hgUser = username
		arg.hgPassword = pwd
	}

	rawByte := fmt.Appendf(nil, "%s:%s", arg.hgUser, arg.hgPassword)
	var dst []byte

	base64.StdEncoding.Encode(dst, rawByte)

	authHeader := fmt.Sprintf("Authorization: Basic %s", string(dst))

	return c.AddMCPServer(MCPAddArg{
		name:      "higress-api",
		url:       fmt.Sprintf("%s/higress-api", arg.baseURL),
		transport: HTTP,
		scope:     "global",
		header: []string{
			authHeader,
		},
	})
}

// ------- Initialization  -------
func (c *AgenticCore) Start() error {
	return c.run(AgentBinaryName)
}

// ------- MCP  -------
func (c *AgenticCore) AddMCPServer(arg MCPAddArg) error {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("mcp add --transport %s %s %s", arg.transport, arg.name, arg.url))
	if arg.scope != "" {
		builder.WriteString(fmt.Sprintf(" --scope %s ", arg.scope))
	}
	if len(arg.env) != 0 {
		str := []string{}
		for _, e := range arg.env {
			str = append(str, fmt.Sprintf("-e %s", e))
		}
		builder.WriteString(strings.Join(str, " "))
	}
	if len(arg.header) != 0 {
		arr := []string{}
		for _, e := range arg.header {
			arr = append(arr, fmt.Sprintf("-h %s", e))
		}
		builder.WriteString(strings.Join(arr, " "))
	}
	cmdStr := builder.String()
	fmt.Println(cmdStr)
	return c.run(cmdStr)
}
