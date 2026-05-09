package report

import (
	"fmt"
	"io"
	"strings"

	"querywise/pkg/types"
)

// WriteMarkdown writes a markdown report to w.
func WriteMarkdown(w io.Writer, rep types.Report) {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# QueryWise Report — %s @ %s\n\n", rep.DatabaseName, rep.HostEndpoint))
	b.WriteString(fmt.Sprintf("**Generated:** %s  \n", rep.GeneratedAt.Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("**Queries analyzed:** %d | **Showing top:** %d  \n\n", rep.QueriesAnalyzed, rep.TopNShown))

	b.WriteString("| Rank | Hash | COST% | Calls | Avg Time | Total Time | Cache Hit | Recommendation |\n")
	b.WriteString("| ---: | :--- | :---: | ---: | :--- | :--- | :--- | :--- |\n")
	for _, row := range rep.Rows {
		rec := row.Recommendation
		if rec == "" {
			rec = ShortHint(row)
		}
		recEscaped := strings.ReplaceAll(rec, "|", "\\|")
		line := fmt.Sprintf(
			"| %d | `%s` | %.1f%% | %s | %s | %s | %.1f%% | %s |\n",
			row.Rank,
			shortHash(row.QueryHash),
			row.CostScore,
			formatInt(row.Calls),
			formatDurationMS(row.MeanExecTimeMs),
			formatTotalTime(row.TotalExecTimeMs),
			row.CacheHitRatio*100,
			recEscaped,
		)
		b.WriteString(line)
	}

	b.WriteString(fmt.Sprintf("\n**Total estimated DB cost covered by top %d:** %.1f%%\n", rep.TopNShown, rep.CostCoveragePct))
	if !rep.LLMUsed {
		b.WriteString("\n_Generate richer guidance with `--recommend`._\n")
	}

	_, _ = w.Write([]byte(b.String()))
}
