package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"querywise/internal/config"
	"querywise/internal/db"
	"querywise/internal/recommender"
	rpt "querywise/internal/report"
	"querywise/internal/scorer"
	"querywise/pkg/types"

	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze pg_stat_statements and emit a ranked report",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}
		return config.Prepare(cmd, cfgPath)
	},
	RunE: runAnalyze,
}

func init() {
	analyzeCmd.Flags().String("config", "", "path to YAML config (.querywise.yml)")
	analyzeCmd.Flags().String("dsn", "", "PostgreSQL DSN")
	analyzeCmd.Flags().Int("top", 10, "number of ranked queries to include")
	analyzeCmd.Flags().String("output", "terminal", "output format: terminal, markdown, or json")
	analyzeCmd.Flags().String("file", "", "output path for markdown or json formats")
	analyzeCmd.Flags().Bool("recommend", false, "request LLM recommendations (requires Anthropic API key)")
	analyzeCmd.Flags().Int64("min-calls", 10, "ignore statements below this calls threshold")

	rootCmd.AddCommand(analyzeCmd)
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	cfg, err := config.ReadAnalyze(cmd)
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	pool, err := db.Connect(ctx, cfg.DSN)
	if err != nil {
		return err
	}
	defer pool.Close()

	stats, err := db.FetchStatements(ctx, pool, cfg.MinCalls)
	if err != nil {
		return err
	}

	dbName, host := db.DSNLabel(cfg.DSN)
	totals := scorer.SumTotals(stats)
	ranked := scorer.RankByScore(stats, totals, cfg.Top)

	rows := make([]types.ScoredQuery, 0, len(ranked))
	for i, r := range ranked {
		hit := float64(r.Stat.SharedBlksHit)
		read := float64(r.Stat.SharedBlksRead)
		ratio := 0.0
		if hit+read > 0 {
			ratio = hit / (hit + read)
		}

		rows = append(rows, types.ScoredQuery{
			Rank:              i + 1,
			QueryHash:         r.Stat.QueryHash,
			CostScore:         r.Score,
			Calls:             r.Stat.Calls,
			MeanExecTimeMs:    r.Stat.MeanExecTimeMs,
			TotalExecTimeMs:   r.Stat.TotalExecTimeMs,
			SharedBlksRead:    r.Stat.SharedBlksRead,
			SharedBlksHit:     r.Stat.SharedBlksHit,
			TempBlksRead:      r.Stat.TempBlksRead,
			TempBlksWritten:   r.Stat.TempBlksWritten,
			Rows:              r.Stat.Rows,
			CacheHitRatio:     ratio,
			Recommendation:    "",
		})
	}

	llmUsed := false
	if cfg.Recommend {
		tips, err := recommender.Recommendations(ctx, cfg.AnthropicAPIKey, cfg.AnthropicModel, ranked)
		if err != nil {
			return err
		}
		llmUsed = true
		for i := range rows {
			key := strings.ToLower(rows[i].QueryHash)
			if msg := strings.TrimSpace(tips[key]); msg != "" {
				rows[i].Recommendation = msg
			}
		}
	}

	rep := types.Report{
		DatabaseName:    dbName,
		HostEndpoint:    host,
		GeneratedAt:     time.Now().UTC(),
		QueriesAnalyzed: len(stats),
		TopNShown:       len(rows),
		CostCoveragePct: scorer.CostCoverage(ranked),
		LLMUsed:         llmUsed,
		Rows:            rows,
	}

	switch cfg.Output {
	case "terminal":
		rpt.WriteTerminal(os.Stdout, rep)
		return nil
	case "markdown":
		f, err := os.Create(cfg.File)
		if err != nil {
			return fmt.Errorf("create markdown file: %w", err)
		}
		defer f.Close()
		rpt.WriteMarkdown(f, rep)
		return nil
	case "json":
		f, err := os.Create(cfg.File)
		if err != nil {
			return fmt.Errorf("create json file: %w", err)
		}
		defer f.Close()
		return rpt.WriteJSON(f, rep)
	default:
		return fmt.Errorf("unknown output format %q", cfg.Output)
	}
}
