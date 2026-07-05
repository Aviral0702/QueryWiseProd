package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Analyze holds resolved settings for the analyze command (flags > env > file > defaults).
type Analyze struct {
	DSN             string
	Top             int
	Output          string
	File            string
	Recommend       bool
	AnthropicAPIKey string
	AnthropicModel  string
	MinCalls        int64
	ScoreTimeWeight float64
	ScoreIOWeight   float64
	ScoreFreqWeight float64
	HashKey         string
}

// Prepare initializes Viper from defaults and optional YAML, then overlays environment variables.
func Prepare(cmd *cobra.Command, cfgFile string) error {
	for _, flag := range []string{"dsn", "top", "output", "file", "recommend", "min-calls", "score-time-weight", "score-io-weight", "score-freq-weight", "hash-key"} {
		if cmd.Flags().Lookup(flag) == nil {
			return fmt.Errorf("missing flag %q", flag)
		}
	}

	viper.Reset()

	viper.SetDefault("top", 10)
	viper.SetDefault("min_calls", int64(10))
	viper.SetDefault("output", "terminal")
	viper.SetDefault("recommend", false)
	viper.SetDefault("file", "")
	viper.SetDefault("dsn", "")
	viper.SetDefault("anthropic_api_key", "")
	viper.SetDefault("anthropic_model", "claude-sonnet-4-20250514")
	viper.SetDefault("score_time_weight", 0.4)
	viper.SetDefault("score_io_weight", 0.4)
	viper.SetDefault("score_freq_weight", 0.2)
	viper.SetDefault("hash_key", "")

	viper.SetEnvPrefix("QUERYWISE")
	viper.AutomaticEnv()
	_ = viper.BindEnv("dsn", "QUERYWISE_DSN")
	_ = viper.BindEnv("anthropic_api_key", "ANTHROPIC_API_KEY")
	_ = viper.BindEnv("anthropic_model", "QUERYWISE_ANTHROPIC_MODEL")
	_ = viper.BindEnv("top", "QUERYWISE_TOP")
	_ = viper.BindEnv("min_calls", "QUERYWISE_MIN_CALLS")
	_ = viper.BindEnv("score_time_weight", "QUERYWISE_SCORE_TIME_WEIGHT")
	_ = viper.BindEnv("score_io_weight", "QUERYWISE_SCORE_IO_WEIGHT")
	_ = viper.BindEnv("score_freq_weight", "QUERYWISE_SCORE_FREQ_WEIGHT")
	_ = viper.BindEnv("hash_key", "QUERYWISE_HASH_KEY")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("read config file: %w", err)
		}
	} else {
		viper.SetConfigName(".querywise")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		if err := viper.ReadInConfig(); err != nil {
			var notFound viper.ConfigFileNotFoundError
			if !errors.As(err, &notFound) {
				return fmt.Errorf("read config: %w", err)
			}
		}
	}

	return nil
}

// ReadAnalyze returns the merged analyze config applying explicit flags last.
func ReadAnalyze(cmd *cobra.Command) (Analyze, error) {
	flags := cmd.Flags()

	out := Analyze{
		DSN:             viper.GetString("dsn"),
		Top:             viper.GetInt("top"),
		Output:          strings.ToLower(strings.TrimSpace(viper.GetString("output"))),
		File:            viper.GetString("file"),
		Recommend:       viper.GetBool("recommend"),
		AnthropicAPIKey: viper.GetString("anthropic_api_key"),
		AnthropicModel:  viper.GetString("anthropic_model"),
		MinCalls:        viper.GetInt64("min_calls"),
		ScoreTimeWeight: viper.GetFloat64("score_time_weight"),
		ScoreIOWeight:   viper.GetFloat64("score_io_weight"),
		ScoreFreqWeight: viper.GetFloat64("score_freq_weight"),
		HashKey:         viper.GetString("hash_key"),
	}

	if fs := flags.Lookup("dsn"); fs != nil && fs.Changed {
		out.DSN = fs.Value.String()
	}
	if fs := flags.Lookup("top"); fs != nil && fs.Changed {
		v, err := flags.GetInt("top")
		if err != nil {
			return Analyze{}, fmt.Errorf("top flag: %w", err)
		}
		out.Top = v
	}
	if fs := flags.Lookup("output"); fs != nil && fs.Changed {
		out.Output = strings.ToLower(strings.TrimSpace(fs.Value.String()))
	}
	if fs := flags.Lookup("file"); fs != nil && fs.Changed {
		out.File = fs.Value.String()
	}
	if fs := flags.Lookup("recommend"); fs != nil && fs.Changed {
		v, err := flags.GetBool("recommend")
		if err != nil {
			return Analyze{}, fmt.Errorf("recommend flag: %w", err)
		}
		out.Recommend = v
	}
	if fs := flags.Lookup("min-calls"); fs != nil && fs.Changed {
		v, err := flags.GetInt64("min-calls")
		if err != nil {
			return Analyze{}, fmt.Errorf("min-calls flag: %w", err)
		}
		out.MinCalls = v
	}
	if fs := flags.Lookup("score-time-weight"); fs != nil && fs.Changed {
		v, err := flags.GetFloat64("score-time-weight")
		if err != nil {
			return Analyze{}, fmt.Errorf("score-time-weight flag: %w", err)
		}
		out.ScoreTimeWeight = v
	}
	if fs := flags.Lookup("score-io-weight"); fs != nil && fs.Changed {
		v, err := flags.GetFloat64("score-io-weight")
		if err != nil {
			return Analyze{}, fmt.Errorf("score-io-weight flag: %w", err)
		}
		out.ScoreIOWeight = v
	}
	if fs := flags.Lookup("score-freq-weight"); fs != nil && fs.Changed {
		v, err := flags.GetFloat64("score-freq-weight")
		if err != nil {
			return Analyze{}, fmt.Errorf("score-freq-weight flag: %w", err)
		}
		out.ScoreFreqWeight = v
	}
	if fs := flags.Lookup("hash-key"); fs != nil && fs.Changed {
		out.HashKey = fs.Value.String()
	}

	if out.DSN == "" {
		return Analyze{}, fmt.Errorf("dsn is required (use --dsn, QUERYWISE_DSN, or config file)")
	}
	if out.Top < 1 {
		return Analyze{}, fmt.Errorf("--top must be >= 1")
	}
	if out.MinCalls < 0 {
		return Analyze{}, fmt.Errorf("--min-calls must be >= 0")
	}
	if out.ScoreTimeWeight < 0 || out.ScoreIOWeight < 0 || out.ScoreFreqWeight < 0 {
		return Analyze{}, fmt.Errorf("score weights must be >= 0")
	}
	if out.ScoreTimeWeight+out.ScoreIOWeight+out.ScoreFreqWeight <= 0 {
		return Analyze{}, fmt.Errorf("score weights must sum to > 0")
	}
	switch out.Output {
	case "terminal", "markdown", "json":
	default:
		return Analyze{}, fmt.Errorf("output must be terminal, markdown, or json")
	}
	if out.Output != "terminal" && strings.TrimSpace(out.File) == "" {
		return Analyze{}, fmt.Errorf("--file is required when --output is %s", out.Output)
	}
	if out.Recommend && strings.TrimSpace(out.AnthropicAPIKey) == "" {
		return Analyze{}, fmt.Errorf("--recommend requires ANTHROPIC_API_KEY or anthropic_api_key in config")
	}

	return out, nil
}
