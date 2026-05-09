package scorer_test

import (
	"math"
	"testing"

	"querywise/internal/db"
	"querywise/internal/scorer"
)

func TestScore(t *testing.T) {
	totals := scorer.Totals{
		TotalExecTime:  100,
		SharedBlksRead: 100,
		Calls:          100,
	}
	q := db.QueryStat{
		TotalExecTimeMs: 40,
		SharedBlksRead:  40,
		Calls:           60,
	}
	got := scorer.Score(q, totals)
	// 0.4*0.4 + 0.4*0.4 + 0.2*0.6 = 0.16+0.16+0.12 = 0.44 * 100 = 44
	want := 44.0
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("Score = %v, want %v", got, want)
	}
}

func TestRankByScore(t *testing.T) {
	stats := []db.QueryStat{
		{QueryHash: "a", TotalExecTimeMs: 10, Calls: 1, SharedBlksRead: 0},
		{QueryHash: "b", TotalExecTimeMs: 90, Calls: 1, SharedBlksRead: 0},
	}
	totals := scorer.SumTotals(stats)
	out := scorer.RankByScore(stats, totals, 1)
	if len(out) != 1 {
		t.Fatalf("len = %d", len(out))
	}
	if out[0].Stat.QueryHash != "b" {
		t.Fatalf("top hash = %q", out[0].Stat.QueryHash)
	}
}
