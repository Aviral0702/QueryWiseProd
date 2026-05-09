package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	rpt "querywise/internal/report"
	"querywise/pkg/types"

	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Load a saved JSON report and render it again",
	RunE:  runSavedReport,
}

func init() {
	reportCmd.Flags().StringP("from", "f", "", "path to JSON report saved by analyze")
	reportCmd.Flags().String("output", "terminal", "output format: terminal, markdown, or json")
	reportCmd.Flags().String("file", "", "output path for markdown or json formats")

	_ = reportCmd.MarkFlagRequired("from")
	rootCmd.AddCommand(reportCmd)
}

func runSavedReport(cmd *cobra.Command, args []string) error {
	path, err := cmd.Flags().GetString("from")
	if err != nil {
		return err
	}

	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}

	outPath, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	output = strings.ToLower(strings.TrimSpace(output))
	switch output {
	case "terminal", "markdown", "json":
	default:
		return fmt.Errorf("output must be terminal, markdown, or json")
	}
	if output != "terminal" && strings.TrimSpace(outPath) == "" {
		return fmt.Errorf("--file is required when --output is %s", output)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read report: %w", err)
	}

	var rep types.Report
	if err := json.Unmarshal(raw, &rep); err != nil {
		return fmt.Errorf("decode JSON report: %w", err)
	}

	switch output {
	case "terminal":
		rpt.WriteTerminal(os.Stdout, rep)
		return nil
	case "markdown":
		f, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("create markdown file: %w", err)
		}
		defer f.Close()
		rpt.WriteMarkdown(f, rep)
		return nil
	default:
		f, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("create json file: %w", err)
		}
		defer f.Close()
		return rpt.WriteJSON(f, rep)
	}
}
