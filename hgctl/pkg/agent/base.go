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

	"github.com/spf13/viper"
)

const (
	AgentBinaryName  = "claude"
	BinaryVersion    = "0.1.0"
	DevVersion       = "dev"
	NodeLeastVersion = 18
	AgentInstallCmd  = "npm install -g @anthropic-ai/claude-code"
	AgentReleasePage = "https://docs.claude.com/en/docs/claude-code/setup"

	HIGRESS_CONSOLE_URL      = "higress-console-url"
	HIGRESS_CONSOLE_USER     = "higress-console-user"
	HIGRESS_CONSOLE_PASSWORD = "higress-console-password"

	HIMARKET_ADMIN_URL      = "himarket-admin-url"
	HIMARKET_ADMIN_USER     = "himarket-admin-user"
	HIMARKET_ADMIN_PASSWORD = "himarket-admin-password"
)

type HimarketAdminAuthArg struct {
	hmURL      string
	hmUser     string
	hmPassword string
}

func (h *HimarketAdminAuthArg) validate() error {
	if h.hmURL == "" || h.hmUser == "" || h.hmPassword == "" {
		return fmt.Errorf("invalid args")
	}
	return nil
}

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

func resolveHimarketAdminAuth(arg *HimarketAdminAuthArg) {
	if arg.hmURL == "" {
		arg.hmURL = viper.GetString(HIMARKET_ADMIN_URL)
	}
	if arg.hmUser == "" {
		arg.hmUser = viper.GetString(HIMARKET_ADMIN_USER)
	}
	if arg.hmPassword == "" {
		arg.hmPassword = viper.GetString(HIMARKET_ADMIN_PASSWORD)
	}
}

// resolve from viper
func resolveHigressConsoleAuth(arg *HigressConsoleAuthArg) {
	if arg.hgURL == "" {
		arg.hgURL = viper.GetString(HIGRESS_CONSOLE_URL)
	}
	if arg.hgUser == "" {
		arg.hgUser = viper.GetString(HIGRESS_CONSOLE_USER)
	}
	if arg.hgPassword == "" {
		arg.hgPassword = viper.GetString(HIGRESS_CONSOLE_PASSWORD)
	}

	// fmt.Printf("arg: %v\n", arg)

	if arg.hgUser == "" || arg.hgPassword == "" {
		// Here we do not return this error, because it will failed when validate arg
		if err := tryToGetLocalCredential(arg); err != nil {
			fmt.Printf("failed to get local higress console credential: %s\n", err)
		}
	}
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
