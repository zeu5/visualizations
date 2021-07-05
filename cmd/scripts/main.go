package main

import (
	"github.com/spf13/cobra"
	datagovin "github.com/zeu5/visualizations/scripts/data.gov.in"
)

func ScriptsRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scripts",
		Short: "run specified script",
	}
	cmd.AddCommand(datagovin.CrimeCmd())
	return cmd
}

func main() {
	cmd := ScriptsRootCmd()
	cmd.Execute()
}
