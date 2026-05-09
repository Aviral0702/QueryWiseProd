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
}

// Prepare initializes Viper from defaults and optional YAML, then overlays environment variables.
func Prepare(cmd *cobra.Command, cfgFile string) error {
	for _, flag := range []string{"dsn", "top", "output", "file", "recommend", "min-calls"} {
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

	viper.SetEnvPrefix("QUERYWISE")
	viper.AutomaticEnv()
	_ = viper.BindEnv("dsn", "QUERYWISE_DSN")
	_ = viper.BindEnv("anthropic_api_key", "ANTHROPIC_API_KEY")
	_ = viper.BindEnv("anthropic_model", "QUERYWISE_ANTHROPIC_MODEL")
	_ = viper.BindEnv("top", "QUERYWISE_TOP")
	_ = viper.BindEnv("min_calls", "QUERYWISE_MIN_CALLS")

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

	if out.DSN == "" {
		return Analyze{}, fmt.Errorf("dsn is required (use --dsn, QUERYWISE_DSN, or config file)")
	}
	if out.Top < 1 {
		return Analyze{}, fmt.Errorf("--top must be >= 1")
	}
	if out.MinCalls < 0 {
		return Analyze{}, fmt.Errorf("--min-calls must be >= 0")
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
