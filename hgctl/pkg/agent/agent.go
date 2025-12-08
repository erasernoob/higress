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

	"github.com/AlecAivazis/survey/v2"
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

type AgentAddArg struct {
	HigressConsoleAuthArg
	HimarketAdminAuthArg

	name  string
	url   string
	typ   string
	scope string

	asProduct bool
	noPublish bool
}

func newAgentAddCmd() *cobra.Command {
	arg := &AgentAddArg{}

	cmd := &cobra.Command{
		Use:   "add [name] [url]",
		Short: "add agent to local interactive window and publish it to higress (optional)",
		Run: func(cmd *cobra.Command, args []string) {
			arg.name = args[0]
			arg.url = args[1]

			resolveHigressConsoleAuth(&arg.HigressConsoleAuthArg)
			resolveHimarketAdminAuth(&arg.HimarketAdminAuthArg)
			cmdutil.CheckErr(handleAddAgent(cmd.OutOrStdout(), *arg))
		},
		Args: cobra.ExactArgs(2),
	}

	cmd.PersistentFlags().StringVarP(&arg.typ, "type", "t", MODEL, "Determine the agent's API type (a2a, model, restful) default is model")
	cmd.PersistentFlags().StringVarP(&arg.scope, "scope", "s", "project", `Configuration scope (project or global)`)
	cmd.PersistentFlags().BoolVar(&arg.noPublish, "no-publish", false, "If it's set then the agent API will not be plubished to Higress")
	cmd.PersistentFlags().BoolVar(&arg.asProduct, "as-product", false, "If it's set then the agent API will be published to Himarket (no-publish must be false)")

	addHigressConsoleAuthFlag(cmd, &arg.HigressConsoleAuthArg)
	return cmd
}

func handleAddAgent(writer io.Writer, arg AgentAddArg) error {
	if err := validateArg(arg); err != nil {
		return err
	}

	if !arg.noPublish {
		if err := publishAgentAPIToHigress(arg); err != nil {
			fmt.Printf("failed to publish agent api to higress: %s\n", err)
			return err
		}

		fmt.Printf("Agent %s is published to Higress successfully\n", arg.name)

		if arg.asProduct {
			if err := publishAgentAPIToHimarket(arg); err != nil {
				fmt.Println("failed to publish it to himarket, please do it mannually")
				return err
			}
			fmt.Printf("Agent %s is published to Himarket successfully\n", arg.name)
		}
		// TODO: pop up higress window
	}

	return nil
}

func publishAgentAPIToHimarket(arg AgentAddArg) error {
	if err := arg.HimarketAdminAuthArg.validate(); err != nil {
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
	var gatewayId string
	prompt := survey.Input{
		Message: "Enter the target Higress instance id on Himarket:",
		Default: "",
		Help:    fmt.Sprintf("refers to %s/consoles/gateway to get your target Higress instance's id", arg.hmURL),
	}

	if err := survey.AskOne(&prompt, &gatewayId); err != nil {
		return fmt.Errorf("failed to get target higress gatewayID: %s", err)
	}

	productName := fmt.Sprintf("agent-%s", arg.name)

	body := services.BuildAPIProductBody(productName, "An agent API import by hgctl", parseTypeToAPIProductType(arg.typ))
	resp, err := services.HandleAddAPIProduct(client, body)
	if err != nil {
		fmt.Println(resp)
		return err
	}

	product_id := string(resp)
	// target_route is the route_name in Higress, refers to `publishAgentAPIToHigress``
	target_route := fmt.Sprintf("%s-route", arg.name)

	body = services.BuildRefAPIProductBody(gatewayId, product_id, target_route)
	if resp, err := services.HandleRefAPIProduct(client, product_id, body); err != nil {
		fmt.Println(resp)
		return err
	}

	return nil
}

func publishAgentAPIToHigress(arg AgentAddArg) error {
	client := services.NewHigressClient(arg.hgURL, arg.hgUser, arg.hgPassword)

	switch arg.typ {
	case A2A:
	case MODEL:
		// add ai service
		body := services.BuildAIProviderServiceBody(arg.name, arg.url)
		if resp, err := services.HandleAddAIProviderService(client, body); err != nil {
			fmt.Println(string(resp))
			return err
		}

		// add ai route
		body = services.BuildAIRouteServiceBody(arg.name, arg.url)
		if res, err := services.HandleAddAIRoute(client, body); err != nil {
			fmt.Println(res)
			return err
		}

	case REST:
		srvName := fmt.Sprintf("agent-%s", arg.name)
		body, targetSrvName, err := services.BuildServiceBodyAndSrvName(srvName, arg.url)
		if err != nil {
			return fmt.Errorf("invalid url format: %s", err)
		}

		if resp, err := services.HandleAddServiceSource(client, body); err != nil {
			fmt.Println(string(resp))
			return err
		}

		if resp, err := services.HandleAddRoute(client, services.BuildAPIRouteBody(arg.name, targetSrvName)); err != nil {
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

func parseTypeToAPIProductType(typ string) string {
	switch typ {
	case "a2a":
		return "AGENT_API"
	case "restful":
		return "REST_API"
	case "model":
		return "MODEL_API"
	default:
		return ""
	}
}
