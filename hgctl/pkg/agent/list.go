package agent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var listAgentsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available agents",
	Run: func(cmd *cobra.Command, args []string) {
		listAgents()
	},
}

func listAgents() {
	fmt.Println()
	color.Cyan("📋 Available Agents:")
	fmt.Println()

	agentsDir := "agents"
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		color.Yellow("No agents found. Create one with 'hgctl new agent <name>'")
		return
	}

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		fmt.Printf("Error reading agents directory: %v\n", err)
		return
	}

	if len(entries) == 0 {
		color.Yellow("No agents found. Create one with 'hgctl new agent <name>'")
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			agentName := entry.Name()
			agentDir := filepath.Join(agentsDir, agentName)
			agentPath := filepath.Join(agentDir, "agent.py")

			if _, err := os.Stat(agentPath); err == nil {
				status := getAgentStatus(agentName)
				fmt.Printf("  🤖 %s - %s\n", agentName, status)
			} else {
				fmt.Printf("  📁 %s - incomplete\n", agentName)
			}
		}
	}
	fmt.Println()
}

func getAgentPort(name string) (int, error) {
	return 8090, nil
}

func getAgentStatus(agentName string) string {
	port, err := getAgentPort(agentName)
	if err != nil {
		return "unknown port"
	}

	if isPortInUse(port) {
		return color.GreenString("running on port " + fmt.Sprintf("%d", port))
	}
	return color.YellowString("stopped")
}
