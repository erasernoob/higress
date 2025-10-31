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

	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

const (
	A2ATransport = "a2a"
)

func NewAgentCmd() *cobra.Command {
	agentCmd := &cobra.Command{
		Use:   "agent",
		Short: "start the interactive agent window",
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(handleAgentInvoke(cmd.OutOrStdout()))
		},
	}

	agentCmd.AddCommand(newAgentAddCmd())

	return agentCmd
}

func handleAgentInvoke(w io.Writer) error {
	return getAgent().Start()
}

// Sub-Agent1:
// 1. Parse the url provided by user to MCP server configuration.
// 2. Publish the parsed MCP Server to Higress
func addPrequisiteSubAgent() error {
	return nil
}

type AgentAddArg struct {
	HigressConsoleAuthArg

	name      string
	url       string
	transport string
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
		Args: cobra.ExactArgs(1),
	}

	cmd.PersistentFlags().StringVarP(&arg.transport, "transport", "t", A2ATransport, "Determine the agent's supported tranport protocol default is A2A")
	cmd.PersistentFlags().StringVarP(&arg.url, "url", "u", "", "Endpoint that the agent serves")
	cmd.PersistentFlags().StringVarP(&arg.scope, "scope", "s", "project", `Configuration scope (project or global)`)
	cmd.PersistentFlags().BoolVar(&arg.noPublish, "no-publish", false, "If set then the agent will not be plubished to higress")

	addHigressConsoleAuthFlag(cmd, &arg.HigressConsoleAuthArg)
	return cmd
}

func handleAddAgent(writer io.Writer, arg AgentAddArg) error {
	agent := getAgent()
	if err := validateArg(arg); err != nil {
		return err
	}

	switch arg.transport {
	case A2ATransport:
		return handleAddA2A(agent, arg)
	default:
		return fmt.Errorf("unsupported agent protocol type: %s", arg.transport)
	}
}

func handleAddA2A(agent *AgenticCore, arg AgentAddArg) error {

	if !arg.noPublish {
		fmt.Println("publish to higress (himarket?)")
	}

	return nil
}

func validateArg(arg AgentAddArg) error {
	if !arg.noPublish {
		return arg.HigressConsoleAuthArg.validate()
	}
	return nil
}
