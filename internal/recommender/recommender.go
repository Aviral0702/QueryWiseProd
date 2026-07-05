package recommender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"querywise/internal/scorer"
	"querywise/pkg/types"
)

const defaultAnthropicURL = "https://api.anthropic.com/v1/messages"

// maxResponseBytes caps how much of the Anthropic HTTP response we read into
// memory, guarding against a compromised or buggy endpoint returning a huge
// body (DoS).
const maxResponseBytes = 5 << 20 // 5 MiB

// readLimited reads at most max bytes from r. If r yields more than max bytes,
// it returns an error instead of buffering the whole body.
func readLimited(r io.Reader, max int64) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > max {
		return nil, fmt.Errorf("anthropic response exceeds %d byte limit", max)
	}
	return b, nil
}

// BuildContexts converts ranked stats into anonymized JSON payloads (no raw SQL).
func BuildContexts(ranked []scorer.Scored) []types.QueryContext {
	out := make([]types.QueryContext, 0, len(ranked))
	for i, r := range ranked {
		hit := float64(r.Stat.SharedBlksHit)
		read := float64(r.Stat.SharedBlksRead)
		ratio := 0.0
		if hit+read > 0 {
			ratio = hit / (hit + read)
		}

		out = append(out, types.QueryContext{
			Rank:            i + 1,
			QueryHash:       r.Stat.QueryHash,
			CostScore:       r.Score,
			Calls:           float64(r.Stat.Calls),
			MeanExecTimeMs:  r.Stat.MeanExecTimeMs,
			TotalExecTimeMs: r.Stat.TotalExecTimeMs,
			SharedBlksRead:  float64(r.Stat.SharedBlksRead),
			SharedBlksHit:   float64(r.Stat.SharedBlksHit),
			TempBlksRead:    float64(r.Stat.TempBlksRead),
			Rows:            float64(r.Stat.Rows),
			CacheHitRatio:   ratio,
		})
	}
	return out
}

type llmPayload struct {
	Items []struct {
		QueryHash      string `json:"query_hash"`
		Recommendation string `json:"recommendation"`
	} `json:"items"`
}

type messagesRequest struct {
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
	Messages  []msg  `json:"messages"`
}
type msg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messagesResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// Recommendations calls the Claude API and returns query_hash -> recommendation text.
func Recommendations(ctx context.Context, apiKey, model string, ranked []scorer.Scored) (map[string]string, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("anthropic api key is empty")
	}
	payload, err := json.MarshalIndent(BuildContexts(ranked), "", "  ")
	if err != nil {
		return nil, err
	}

	prompt := strings.TrimSpace(`
You are a PostgreSQL performance expert. Given these anonymized query stats
(no raw SQL — only hashed identifiers and numeric metrics), identify likely
performance problems and suggest optimizations. Be specific about what the
numbers indicate (e.g., low cache hit ratio may suggest sequential scans).

Respond with JSON ONLY in this shape (no markdown fences):
{"items":[{"query_hash":"<hex>","recommendation":"<one short actionable sentence>"}]}

Stats JSON:
`) + "\n" + string(payload)

	reqBody := messagesRequest{
		Model:     model,
		MaxTokens: 4096,
		Messages:  []msg{{Role: "user", Content: prompt}},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, defaultAnthropicURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := readLimited(resp.Body, maxResponseBytes)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("anthropic api error: %s — %s", resp.Status, string(respBody))
	}

	var parsed messagesResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("decode anthropic response: %w", err)
	}
	if len(parsed.Content) == 0 {
		return nil, fmt.Errorf("empty anthropic content")
	}
	var text string
	for _, block := range parsed.Content {
		if strings.EqualFold(block.Type, "text") && strings.TrimSpace(block.Text) != "" {
			text = block.Text
			break
		}
	}
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("no text blocks in anthropic response")
	}
	for strings.HasPrefix(strings.TrimSpace(text), "```") {
		text = stripFence(text)
	}

	var lp llmPayload
	if err := json.Unmarshal([]byte(text), &lp); err != nil {
		return nil, fmt.Errorf("decode llm json: %w — raw: %s", err, truncate(text))
	}

	out := make(map[string]string)
	for _, it := range lp.Items {
		if it.QueryHash == "" {
			continue
		}
		out[strings.ToLower(it.QueryHash)] = strings.TrimSpace(it.Recommendation)
	}
	return out, nil
}

func stripFence(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func truncate(s string) string {
	if len(s) > 800 {
		return s[:800] + "…"
	}
	return s
}
