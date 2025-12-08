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
	"fmt"
	"os"
)

const (
	AgentBinaryName  = "claude"
	BinaryVersion    = "0.1.0"
	DevVersion       = "dev"
	NodeLeastVersion = 18
	AgentInstallCmd  = "npm install -g @anthropic-ai/claude-code"
	AgentReleasePage = "https://docs.claude.com/en/docs/claude-code/setup"
)

type HigressConsoleAuthArg struct {
	// higress console auth arg
	hgURL      string
	hgUser     string
	hgPassword string
}

func (h *HigressConsoleAuthArg) validate() error {
	if h.hgURL == "" || h.hgUser == "" || h.hgPassword == "" {
		fmt.Println("--higress-console-user, --higress-console-url, --higress-console-password must be provided")
		return fmt.Errorf("invalid args")
	}
	return nil
}

// set up the core env
// 1. check if npm is installed
// 2. check the npm version
// 3. install hgctl-agent
func getAgent() *AgenticCore {
	if !checkAgentInstallStatus() {
		fmt.Println("⚠️ Prerequisites not satisfied. Exiting...")
		// exit directly
		os.Exit(1)
	}

	return NewAgenticCore()
}

func checkAgentInstallStatus() bool {
	// TODO: Support cross-platform:windows

	if !checkNodeInstall() {
		if err := promptNodeInstall(); err != nil {
			return false
		}
	}

	if !checkAgentInstall() {
		if err := promptAgentInstall(); err != nil {
			return false
		}
	}

	return true
}
