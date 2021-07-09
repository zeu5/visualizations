package datagovin

import "github.com/spf13/cobra"

func CrimeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crime",
		Short: "Fetch/dump Crime records from data.gov.in",
	}
	cmd.PersistentFlags().StringVar(&dbURL, "mongo", "mongodb://localhost:27017", "MongoDB URI")

	fetchCmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch Crime records from data.gov.in",
		Run: func(cmd *cobra.Command, args []string) {
			Fetch()
		},
	}
	dumpCmd := &cobra.Command{
		Use:   "dump",
		Short: "Dump data in mongo as jsonl",
		Run: func(cmd *cobra.Command, args []string) {
			Dump()
		},
	}
	summaryCmd := &cobra.Command{
		Use:   "summary",
		Short: "Summarize all data",
		Run: func(cmd *cobra.Command, args []string) {
			Summarise()
		},
	}
	dumpCmd.PersistentFlags().StringVar(&dumpPath, "path", "dump", "Path to dump data at")

	cmd.AddCommand(fetchCmd)
	cmd.AddCommand(dumpCmd)
	cmd.AddCommand(summaryCmd)
	return cmd
}
