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
	"io"
	"net"
	"net/url"

	"github.com/alibaba/higress/hgctl/pkg/agent/services"
	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// API Type
const (
	A2A   = "a2a"
	REST  = "restful"
	MODEL = "model"
)

func NewAgentCmd() *cobra.Command {
	agentCmd := &cobra.Command{
		Use:   "agent",
		Short: "start the interactive agent window",
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(handleAgentInvoke(cmd.OutOrStdout()))
		},
	}

	agentCmd.AddCommand(createAgentCmd())
	agentCmd.AddCommand(newAgentAddCmd())

	return agentCmd
}

func handleAgentInvoke(w io.Writer) error {
	return getAgent().Start()
}

type Publisher struct {
	inner services.HigressClient
}

type AgentAddArg struct {
	HigressConsoleAuthArg

	name      string
	url       string
	typ       string
	scope     string
	noPublish bool
}

func newAgentAddCmd() *cobra.Command {
	// parameter
	arg := &AgentAddArg{}

	cmd := &cobra.Command{
		Use:   "add [name] [url]",
		Short: "add agent to local interactive window and publish it to higress (optional)",
		Run: func(cmd *cobra.Command, args []string) {
			arg.name = args[0]
			arg.url = args[1]

			resolveHigressConsoleAuth(&arg.HigressConsoleAuthArg)
			cmdutil.CheckErr(handleAddAgent(cmd.OutOrStdout(), *arg))
		},
		Args: cobra.ExactArgs(2),
	}

	cmd.PersistentFlags().StringVarP(&arg.typ, "type", "t", MODEL, "Determine the agent's supported tranport protocol default is A2A")
	cmd.PersistentFlags().StringVarP(&arg.scope, "scope", "s", "project", `Configuration scope (project or global)`)
	cmd.PersistentFlags().BoolVar(&arg.noPublish, "no-publish", false, "If set then the agent will not be plubished to higress")

	addHigressConsoleAuthFlag(cmd, &arg.HigressConsoleAuthArg)
	return cmd
}

func handleAddAgent(writer io.Writer, arg AgentAddArg) error {
	// agent := getAgent()
	if err := validateArg(arg); err != nil {
		return err
	}

	if err := publishAgentEndpointToHigress(arg); err != nil {
		fmt.Printf("failed to publish agent api to higress: %s\n", err)
		return err
	}

	return nil
}

func (p *Publisher) addAIProvider(body interface{}) {

}

func publishAgentEndpointToHigress(arg AgentAddArg) error {
	client := services.NewHigressClient(arg.hgURL, arg.hgUser, arg.hgPassword)
	// publisher := Publisher{inner: *client}

	switch arg.typ {
	case A2A:
	case MODEL:
		// add ai service
		customBaseURL := fmt.Sprintf("%s/compatible-mode/v1", arg.url)
		body := map[string]interface{}{
			"type":     "openai",
			"name":     arg.name,
			"tokens":   []string{},
			"version":  0,
			"protocol": "openai/v1",
			"tokenFailoverConfig": map[string]interface{}{
				"enabled": false,
			},
			"proxyName": "",
			"rawConfigs": map[string]interface{}{
				"openaiExtraCustomUrls": []string{},
				"openaiCustomUrl":       customBaseURL,
			},
		}
		if resp, err := services.HandleAddAIProviderService(client, body); err != nil {
			fmt.Println(string(resp))
			return err
		}

		// add ai route
		if res, err := services.HandleAddAIRoute(client, map[string]interface{}{
			"name": fmt.Sprintf("%s-api", arg.name),
			// "version": "627198", // 创建时不需提供
			"domains": []interface{}{},
			"pathPredicate": map[string]interface{}{
				"matchType":     "PRE",
				"matchValue":    "/",
				"caseSensitive": false,
			},
			"headerPredicates":   []interface{}{},
			"urlParamPredicates": []interface{}{},
			"upstreams": []interface{}{
				map[string]interface{}{
					"provider":     arg.name,
					"weight":       100,
					"modelMapping": map[string]interface{}{},
				},
			},
			"modelPredicates": []interface{}{},
			"authConfig": map[string]interface{}{
				"enabled":                false,
				"allowedCredentialTypes": nil,
				"allowedConsumers":       []interface{}{},
			},
			"fallbackConfig": map[string]interface{}{
				"enabled":          false,
				"upstreams":        nil,
				"fallbackStrategy": nil,
				"responseCodes":    nil,
			},
		}); err != nil {
			fmt.Println(res)
			return err
		}

	case REST:
		res, err := url.Parse(arg.url)
		if err != nil {
			return err
		}

		// add service source
		srvType := ""
		srvPort := ""
		srvName := fmt.Sprintf("agent-%s", arg.name)

		if ip := net.ParseIP(res.Hostname()); ip == nil {
			srvType = "dns"
		} else {
			srvType = "static"
		}

		if res.Port() == "" && res.Scheme == "http" {
			srvPort = "80"
		} else if res.Port() == "" && res.Scheme == "https" {
			srvPort = "443"
		} else {
			srvPort = res.Port()
		}

		if resp, err := services.HandleAddServiceSource(client, map[string]interface{}{
			"domain":        res.Host,
			"type":          srvType,
			"port":          srvPort,
			"name":          srvName,
			"domainForEdit": res.Host,
			"protocol":      res.Scheme,
		}); err != nil {
			fmt.Println(string(resp))
			return err
		}

		// e.g. agent-jarvis.static.8090
		targetSrvName := fmt.Sprintf("%s.%s:%s", srvName, srvType, srvPort)

		// add route
		if resp, err := services.HandleAddRoute(client, map[string]interface{}{
			"name": arg.name,
			"path": map[string]interface{}{
				"matchType":     "PRE",      // default is PREFIX
				"matchValue":    "/process", // default is "/process"
				"caseSensitive": true,
			},
			"authConfig": map[string]interface{}{
				"enabled": false,
			},
			"services": []map[string]interface{}{
				{
					"name": targetSrvName,
				},
			},
		}); err != nil {
			fmt.Println(string(resp))
			return err
		}

	default:
		return fmt.Errorf("unsupported agent protocol type: %s", arg.typ)

	}

	return nil
}

func validateArg(arg AgentAddArg) error {
	if !arg.noPublish {
		return arg.HigressConsoleAuthArg.validate()
	}
	return nil
}
