package main

import (
	"github.com/spf13/cobra"
	"github.com/zeu5/visualizations/server"
)

var (
	configPath string
)

func ServerRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Run the server",
		Run: func(cmd *cobra.Command, args []string) {
			server.Run(configPath)
		},
	}
	cmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to the config file")
	return cmd
}

func main() {
	cmd := ServerRootCmd()
	cmd.Execute()
}
