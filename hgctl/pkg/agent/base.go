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

	"github.com/AlecAivazis/survey/v2"
	"github.com/alibaba/higress/hgctl/pkg/agent/services"
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

func parseTypeToAPIProductType(typ string) string {
	switch typ {
	case "a2a":
		return "AGENT_API"
	case "restful":
		return "REST_API"
	case "model":
		return "MODEL_API"
	case "mcp":
		return "MCP_SERVER"
	default:
		return ""
	}
}

// This function serves MCP API as well as Model API for now.
func publishAPIToHimarket(typ, name string, arg HimarketAdminAuthArg) error {

	if err := arg.validate(); err != nil {
		return err
	}

	// hgName := "hgctl-higress"
	// hgAddress := arg.hgURL
	// hgUsername := arg.hgUser
	// hgPassword := arg.hgPassword

	client := services.NewHimarketClient(arg.hmURL, arg.hmUser, arg.hmPassword)
	// if resp, err := services.HandleAddHigressInstance(client, services.BuildAddHigressInstanceBody(hgName, hgAddress, hgUsername, hgPassword)); err != nil {
	// 	fmt.Println(string(resp))
	// 	return err
	// }

	productName := fmt.Sprintf("%s-%s", typ, name)

	var gatewayId string
	prompt := survey.Input{
		Message: "Enter the target Higress instance id on Himarket:",
		Default: "",
		Help:    fmt.Sprintf("refers to %s/consoles/gateway to get your target Higress instance's id", arg.hmURL),
	}

	if err := survey.AskOne(&prompt, &gatewayId); err != nil {
		return fmt.Errorf("failed to get target higress gatewayID: %s", err)
	}

	body := services.BuildAPIProductBody(productName, "An agent API import by hgctl", parseTypeToAPIProductType(typ))
	resp, err := services.HandleAddAPIProduct(client, body)
	if err != nil {
		fmt.Println(resp)
		return err
	}

	product_id := string(resp)
	var refBody map[string]interface{}

	if typ == "mcp" {
		refBody = services.BuildRefMCPAPIProductBody(gatewayId, product_id, name)
	} else {
		// target_route is the route_name in Higress, refers to `publishAgentAPIToHigress`
		target_route := fmt.Sprintf("%s-route", name)
		refBody = services.BuildRefModelAPIProductBody(gatewayId, product_id, target_route)

	}

	if resp, err := services.HandleRefAPIProduct(client, product_id, refBody); err != nil {
		fmt.Println(string(resp))
		return err
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
