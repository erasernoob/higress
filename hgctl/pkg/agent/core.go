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
	// Check if this is the first time, otherwise directly return (this is a simple check)
	homeDir, _ := os.UserHomeDir()
	targetDir := filepath.Join(homeDir, ".hgctl")
	if _, err := os.Stat(targetDir); err == nil {
		return
	}

	// Setup subagent plugins file
	embedFS := manifests.BuiltinOrDir("")
	if err := manifests.ExtractEmbedFiles(embedFS, "agent", targetDir); err != nil {
		fmt.Println(err)
		fmt.Println("failed to init plugins for claude code")
		os.Exit(1)
	}

	if err := c.addHigressAPIMCP(); err != nil {
		fmt.Println("failed to init higress-api mcp server (you may need to add it manually):", err)
		fmt.Println("Details information on Higress-api MCP server refers to https://github.com/alibaba/higress/blob/main/plugins/golang-filter/mcp-server/servers/higress/higress-api/README_en.md")
		return
	}
}

func (c *AgenticCore) addHigressAPIMCP() error {
	arg := &HigressConsoleAuthArg{
		hgURL:      "",
		hgUser:     "",
		hgPassword: "",
	}
	fmt.Println("Initializing...Add prequisite MCP server (Higress-api MCP server) automatically")
	gatewayPrompt := promptui.Prompt{
		Label:   "Enter higress gateway URL",
		Default: "http://127.0.0.1:80",
	}
	gateway, err := gatewayPrompt.Run()
	if err != nil {
		fmt.Println("failed to run gateway prompt: ", err)
	}

	arg.hgURL = gateway

	if err := tryToGetLocalCredential(arg); err != nil || arg.hgUser == "" || arg.hgPassword == "" {
		// fallback: interact with user to provide password & username
		color.Red("failed to get higress-console credential automatically (Need install higress by hgctl). Let's fix it manually")
		userPrompt := promptui.Prompt{
			Label:   "Enter higress console username",
			Default: "admin",
		}
		username, err := userPrompt.Run()
		if err != nil {
			return fmt.Errorf("aborted: %v", err)
		}
		pwdPrompt := promptui.Prompt{
			Label:   "Enter higress console password",
			Default: "admin",
		}
		pwd, err := pwdPrompt.Run()
		if err != nil {
			return fmt.Errorf("aborted: %v", err)
		}
		arg.hgUser = username
		arg.hgPassword = pwd
	}

	if arg.hgUser == "" || arg.hgPassword == "" {
		return fmt.Errorf("Empty higress console username and password, aborting")
	}

	rawByte := fmt.Appendf(nil, "%s:%s", arg.hgUser, arg.hgPassword)

	resStr := base64.StdEncoding.EncodeToString(rawByte)

	authHeader := fmt.Sprintf("Authorization: Basic %s", resStr)

	return c.AddMCPServer(MCPAddArg{
		name:  "higress-api",
		url:   fmt.Sprintf("%s/higress-api", arg.hgURL),
		typ:   HTTP,
		scope: "user",
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
	// adapt the field
	if arg.transport == STREAMABLE {
		arg.transport = HTTP
	}
	args := []string{
		"mcp", "add", "--transport", arg.transport, arg.name, arg.url,
	}
	if arg.scope != "" {
		scopeArg := []string{"--scope", arg.scope}
		args = append(args, scopeArg...)
	}
	if len(arg.env) != 0 {
		for _, e := range arg.env {
			envArg := []string{"-e", e}
			args = append(args, envArg...)
		}
	}
	if len(arg.header) != 0 {
		for _, h := range arg.header {
			headerArg := []string{"-H", h}
			args = append(args, headerArg...)
		}
	}
	return c.run(args...)
}
