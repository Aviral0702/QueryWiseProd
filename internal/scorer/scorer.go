package scorer

import (
	"math"
	"sort"

	"querywise/internal/db"
)

// Totals aggregates fields used for cost shares.
type Totals struct {
	TotalExecTime  float64
	SharedBlksRead float64
	Calls          float64
}

// SumTotals computes totals across all statements.
func SumTotals(stats []db.QueryStat) Totals {
	var t Totals
	for _, q := range stats {
		t.TotalExecTime += q.TotalExecTimeMs
		t.SharedBlksRead += float64(q.SharedBlksRead)
		t.Calls += float64(q.Calls)
	}
	return t
}

// Weights controls the relative contribution of each cost share to the score.
type Weights struct {
	Time float64
	IO   float64
	Freq float64
}

// DefaultWeights is the standard weighting used when none is configured.
var DefaultWeights = Weights{Time: 0.4, IO: 0.4, Freq: 0.2}

// Score returns a 0–100 cost share for this query pattern given database-wide totals.
//
// The score is a weighted blend of three shares:
//
//	score = (timeShare*wTime + ioShare*wIO + freqShare*wFreq) * 100
//
// where the weights are normalized by their sum so the result stays a
// percentage regardless of the raw weight values. Zero or negative weight
// sums fall back to DefaultWeights.
func Score(q db.QueryStat, totals Totals, w Weights) float64 {
	sum := w.Time + w.IO + w.Freq
	if sum <= 0 {
		w = DefaultWeights
		sum = w.Time + w.IO + w.Freq
	}
	wTime := w.Time / sum
	wIO := w.IO / sum
	wFreq := w.Freq / sum

	var timeShare, ioShare, freqShare float64

	if totals.TotalExecTime > 0 {
		timeShare = q.TotalExecTimeMs / totals.TotalExecTime
	}
	if totals.SharedBlksRead > 0 {
		ioShare = float64(q.SharedBlksRead) / totals.SharedBlksRead
	}
	if totals.Calls > 0 {
		freqShare = float64(q.Calls) / totals.Calls
	}

	return (timeShare*wTime + ioShare*wIO + freqShare*wFreq) * 100
}

// Scored pairs a stat with its score (not yet ranked).
type Scored struct {
	Stat  db.QueryStat
	Score float64
}

// RankByScore sorts by descending score and returns at most topN items.
func RankByScore(stats []db.QueryStat, totals Totals, w Weights, topN int) []Scored {
	scored := make([]Scored, 0, len(stats))
	for _, q := range stats {
		scored = append(scored, Scored{Stat: q, Score: Score(q, totals, w)})
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].Score == scored[j].Score {
			return scored[i].Stat.TotalExecTimeMs > scored[j].Stat.TotalExecTimeMs
		}
		return scored[i].Score > scored[j].Score
	})

	if topN > 0 && len(scored) > topN {
		scored = scored[:topN]
	}
	return scored
}

// CostCoverage returns the sum of cost scores for the provided ranked slice.
func CostCoverage(ranked []Scored) float64 {
	var sum float64
	for _, r := range ranked {
		sum += r.Score
	}
	// Cap at 100 for floating noise
	return math.Min(sum, 100)
}
