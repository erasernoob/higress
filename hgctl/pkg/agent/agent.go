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
	if err := validateArg(arg); err != nil {
		return err
	}

	if err := publishAgentEndpointToHigress(arg); err != nil {
		fmt.Printf("failed to publish agent api to higress: %s\n", err)
		return err
	}

	return nil
}

func publishAgentEndpointToHigress(arg AgentAddArg) error {
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
