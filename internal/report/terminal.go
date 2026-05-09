package report

import (
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"querywise/pkg/types"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

// WriteTerminal renders a colored table to w.
func WriteTerminal(w io.Writer, rep types.Report) {
	title := fmt.Sprintf("QueryWise Report — %s @ %s", rep.DatabaseName, rep.HostEndpoint)
	cyan := color.New(color.FgCyan).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	_, _ = fmt.Fprintf(w, "%s\n", bold(title))
	_, _ = fmt.Fprintf(w, "Generated: %s\n", rep.GeneratedAt.Format("2006-01-02 15:04:05"))
	_, _ = fmt.Fprintf(w, "%s\n\n", cyan(fmt.Sprintf("Queries analyzed: %d | Showing top %d", rep.QueriesAnalyzed, rep.TopNShown)))

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"RANK", "HASH", "COST%", "CALLS", "AVG TIME", "TOTAL TIME", "CACHE HIT", "RECOMMENDATION"})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("─")
	table.SetColumnSeparator(" ")
	table.SetRowSeparator("─")
	table.SetBorder(false)
	table.SetTablePadding(" ")
	table.SetNoWhiteSpace(true)

	for _, row := range rep.Rows {
		rec := row.Recommendation
		if rec == "" {
			rec = ShortHint(row)
		}
		table.Append([]string{
			fmt.Sprintf("%d", row.Rank),
			shortHash(row.QueryHash),
			fmt.Sprintf("%.1f%%", row.CostScore),
			formatInt(row.Calls),
			formatDurationMS(row.MeanExecTimeMs),
			formatTotalTime(row.TotalExecTimeMs),
			fmt.Sprintf("%.1f%%", row.CacheHitRatio*100),
			rec,
		})
	}

	table.Render()

	_, _ = fmt.Fprintf(w, "\nTotal estimated DB cost covered by top %d: %.1f%%\n", rep.TopNShown, rep.CostCoveragePct)
	if !rep.LLMUsed {
		_, _ = fmt.Fprintln(w, "\nRun with --recommend for detailed LLM-powered recommendations.")
	}
}

func shortHash(hexFull string) string {
	if len(hexFull) <= 8 {
		return hexFull
	}
	return hexFull[:8]
}

func formatInt(v int64) string {
	s := fmt.Sprintf("%d", v)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func formatDurationMS(ms float64) string {
	if ms < 1 {
		return fmt.Sprintf("%.2fms", ms)
	}
	if ms < 10 {
		return fmt.Sprintf("%.1fms", ms)
	}
	return fmt.Sprintf("%.0fms", ms)
}

func formatTotalTime(ms float64) string {
	sec := ms / 1000
	if sec < 60 {
		return fmt.Sprintf("%ss", formatInt(int64(math.Round(sec))))
	}
	d := time.Duration(sec * float64(time.Second))
	return d.Round(time.Second).String()
}
