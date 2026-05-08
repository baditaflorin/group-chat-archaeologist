package config

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	InputPath   string
	OutputDir   string
	TikaURL     string
	OllamaURL   string
	OllamaModel string
	Start       time.Time
	End         time.Time
	Concurrency int
	SaveEvery   int
}

func Load(args []string) (Config, error) {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetDefault("INPUT_PATH", "./testdata/sample_chat.txt")
	viper.SetDefault("OUTPUT_DIR", "./docs/data/v1")
	viper.SetDefault("OLLAMA_MODEL", "llama3.2")
	viper.SetDefault("CONCURRENCY", 4)
	viper.SetDefault("SAVE_EVERY", 5000)

	cfg := Config{}
	fs := flag.NewFlagSet("build-index", flag.ContinueOnError)
	fs.StringVar(&cfg.InputPath, "input_path", viper.GetString("INPUT_PATH"), "chat export path")
	fs.StringVar(&cfg.OutputDir, "output_dir", viper.GetString("OUTPUT_DIR"), "artifact output directory")
	fs.StringVar(&cfg.TikaURL, "tika_url", viper.GetString("TIKA_SERVER_URL"), "Apache Tika server URL")
	fs.StringVar(&cfg.OllamaURL, "ollama_url", viper.GetString("OLLAMA_URL"), "Ollama-compatible local LLM URL")
	fs.StringVar(&cfg.OllamaModel, "ollama_model", viper.GetString("OLLAMA_MODEL"), "local LLM model")
	fs.IntVar(&cfg.Concurrency, "concurrency", viper.GetInt("CONCURRENCY"), "batch concurrency")
	fs.IntVar(&cfg.SaveEvery, "saveEvery", viper.GetInt("SAVE_EVERY"), "checkpoint interval")
	fs.IntVar(&cfg.SaveEvery, "save_every", viper.GetInt("SAVE_EVERY"), "checkpoint interval")

	start := fs.String("start", viper.GetString("START"), "inclusive start date YYYY-MM-DD")
	end := fs.String("end", viper.GetString("END"), "inclusive end date YYYY-MM-DD")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	if cfg.InputPath == "" {
		return Config{}, fmt.Errorf("input_path is required")
	}
	if cfg.OutputDir == "" {
		return Config{}, fmt.Errorf("output_dir is required")
	}
	if cfg.Concurrency < 1 {
		return Config{}, fmt.Errorf("concurrency must be >= 1")
	}
	if cfg.SaveEvery < 1 {
		return Config{}, fmt.Errorf("saveEvery must be >= 1")
	}

	parsedStart, err := parseBoundary(*start, false)
	if err != nil {
		return Config{}, err
	}
	parsedEnd, err := parseBoundary(*end, true)
	if err != nil {
		return Config{}, err
	}
	cfg.Start = parsedStart
	cfg.End = parsedEnd

	return cfg, nil
}

func parseBoundary(value string, endOfDay bool) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse boundary %q: %w", value, err)
	}
	if endOfDay {
		return parsed.Add(24*time.Hour - time.Nanosecond), nil
	}
	return parsed, nil
}
