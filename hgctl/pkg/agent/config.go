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
	"log"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type AgentCore string

const (
	CLAUDE_CODE AgentCore = "claude"
	QODER_CLI   AgentCore = "qodercli"
)

const (
	// AgentBinaryName  = "claude"
	// BinaryVersion    = "0.1.0"
	// DevVersion       = "dev"
	// NodeLeastVersion = 18
	// AgentInstallCmd  = "npm install -g @anthropic-ai/claude-code"
	// AgentReleasePage = "https://docs.claude.com/en/docs/claude-code/setup"

	HGCTL_AGENT_CORE         = "hgctl-agent-core"
	AGENT_CHAT_MODEL         = "agent-chat-model"
	HIGRESS_CONSOLE_URL      = "higress-console-url"
	HIGRESS_CONSOLE_USER     = "higress-console-user"
	HIGRESS_CONSOLE_PASSWORD = "higress-console-password"
	HIGRESS_GATEWAY_URL      = "higress-gateway-url"
	HIMARKET_ADMIN_URL       = "himarket-admin-url"
	HIMARKET_ADMIN_USER      = "himarket-admin-user"
	HIMARKET_ADMIN_PASSWORD  = "himarket-admin-password"

	// --- AgentRun ---
	AGENTRUN_MODEL_NAME             = "agentrun-model-name"
	AGENTRUN_SANDBOX_NAME           = "agentrun-sandbox-name"
	ALIBABA_CLOUD_ACCESS_KEY_ID     = "alibaba-cloud-access-key-id"
	ALIBABA_CLOUD_ACCESS_KEY_SECRET = "alibaba-cloud-access-key-secret"
	ALIBABA_CLOUD_SECURITY_TOK      = "alibaba-cloud-security-tok"
	AGENTRUN_ACCOUNT_ID             = "agentrun-account-id"
	AGENTRUN_REGION                 = "agentrun-region"
	AGENTRUN_SDK_DEB                = "agentrun-sdk-deb"
)

var GlobalConfig HgctlAgentConfig

type HgctlAgentConfig struct {
	AGENT_CORE AgentCore `mapstructure:"hgctl-agent-core"`

	// Higress Console credentials
	HigressConsoleURL      string `mapstructure:"higress-console-url"`
	HigressConsoleUser     string `mapstructure:"higress-console-user"`
	HigressConsolePassword string `mapstructure:"higress-console-password"`
	HigressGatewayURL      string `mapstructure:"higress-gateway-url"`
	// Himarket Admin credentials
	HimarketAdminURL      string `mapstructure:"himarket-admin-url"`
	HimarketAdminUser     string `mapstructure:"himarket-admin-user"`
	HimarketAdminPassword string `mapstructure:"himarket-admin-password"`

	// --- AgentRun Configuration ---
	AgentRunModelName           string `mapstructure:"agentrun-model-name"`
	AgentRunSandboxName         string `mapstructure:"agentrun-sandbox-name"`
	AlibabaCloudAccessKeyID     string `mapstructure:"alibaba-cloud-access-key-id"`
	AlibabaCloudAccessKeySecret string `mapstructure:"alibaba-cloud-access-key-secret"`
	AlibabaCloudSecurityTok     string `mapstructure:"alibaba-cloud-security-tok"`
	AgentRunAccountID           string `mapstructure:"agentrun-account-id"`
	AgentRunRegion              string `mapstructure:"agentrun-region"`
}

func InitConfig() {
	viper.SetConfigName(".hgctl")
	viper.SetConfigType("json")

	home, err := homedir.Dir()
	if err != nil {
		log.Fatalf("Error finding home directory: %v", err)
	}

	viper.AddConfigPath(home)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Fatal error reading config file: %v\n", err)
		}
	}

	// Unmarshal into the GlobalConfig variable
	_ = viper.Unmarshal(&GlobalConfig)

	switch viper.GetString(HGCTL_AGENT_CORE) {
	case string(CLAUDE_CODE), string(QODER_CLI):
		return
	default:
		viper.SetDefault(HGCTL_AGENT_CORE, string(CLAUDE_CODE))
	}

}
