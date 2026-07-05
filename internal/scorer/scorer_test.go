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
	got := scorer.Score(q, totals, scorer.DefaultWeights)
	// 0.4*0.4 + 0.4*0.4 + 0.2*0.6 = 0.16+0.16+0.12 = 0.44 * 100 = 44
	want := 44.0
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("Score = %v, want %v", got, want)
	}
}

func TestScoreZeroSumWeightsFallsBackToDefault(t *testing.T) {
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
	got := scorer.Score(q, totals, scorer.Weights{})
	want := scorer.Score(q, totals, scorer.DefaultWeights)
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("zero-sum weights Score = %v, want default %v", got, want)
	}
}

func TestScoreCustomWeightsChangeResult(t *testing.T) {
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
	// Weighting frequency exclusively should yield freqShare*100 = 60.
	got := scorer.Score(q, totals, scorer.Weights{Freq: 1})
	want := 60.0
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("freq-only Score = %v, want %v", got, want)
	}
}

func TestRankByScore(t *testing.T) {
	stats := []db.QueryStat{
		{QueryHash: "a", TotalExecTimeMs: 10, Calls: 1, SharedBlksRead: 0},
		{QueryHash: "b", TotalExecTimeMs: 90, Calls: 1, SharedBlksRead: 0},
	}
	totals := scorer.SumTotals(stats)
	out := scorer.RankByScore(stats, totals, scorer.DefaultWeights, 1)
	if len(out) != 1 {
		t.Fatalf("len = %d", len(out))
	}
	if out[0].Stat.QueryHash != "b" {
		t.Fatalf("top hash = %q", out[0].Stat.QueryHash)
	}
}

func TestRankByScoreCustomWeightsChangeRanking(t *testing.T) {
	stats := []db.QueryStat{
		// "a" dominates call frequency; "b" dominates exec time.
		{QueryHash: "a", TotalExecTimeMs: 10, Calls: 90, SharedBlksRead: 0},
		{QueryHash: "b", TotalExecTimeMs: 90, Calls: 10, SharedBlksRead: 0},
	}
	totals := scorer.SumTotals(stats)

	timeTop := scorer.RankByScore(stats, totals, scorer.Weights{Time: 1}, 1)
	if timeTop[0].Stat.QueryHash != "b" {
		t.Fatalf("time-weighted top = %q, want b", timeTop[0].Stat.QueryHash)
	}

	freqTop := scorer.RankByScore(stats, totals, scorer.Weights{Freq: 1}, 1)
	if freqTop[0].Stat.QueryHash != "a" {
		t.Fatalf("freq-weighted top = %q, want a", freqTop[0].Stat.QueryHash)
	}
}
