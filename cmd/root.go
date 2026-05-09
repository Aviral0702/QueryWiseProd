package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "querywise",
	Short: "Analyze PostgreSQL pg_stat_statements with anonymized hashing and ranking",
}

// Execute runs the CLI root command.
func Execute() error {
	return rootCmd.Execute()
}
