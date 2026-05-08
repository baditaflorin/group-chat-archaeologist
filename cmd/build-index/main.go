package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/baditaflorin/group-chat-archaeologist/internal/analyze"
	"github.com/baditaflorin/group-chat-archaeologist/internal/artifact"
	"github.com/baditaflorin/group-chat-archaeologist/internal/chatparse"
	"github.com/baditaflorin/group-chat-archaeologist/internal/config"
	"github.com/baditaflorin/group-chat-archaeologist/internal/extract"
	"github.com/baditaflorin/group-chat-archaeologist/internal/storage"
	"github.com/baditaflorin/group-chat-archaeologist/internal/utils"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	cfg, err := config.Load(os.Args[1:])
	utils.HandleErrorOrLogWithMessages(err, "failed to load config", "")

	text, extractionMode, err := extract.Text(ctx, cfg.InputPath, cfg.TikaURL)
	utils.HandleErrorOrLogWithMessages(err, "failed to extract chat text", "")

	messages, parserName, err := chatparse.Parse(text)
	utils.HandleErrorOrLogWithMessages(err, "failed to parse chat export", "")
	messages = chatparse.Filter(messages, cfg.Start, cfg.End)

	store, err := storage.Open()
	utils.HandleErrorOrLogWithMessages(err, "failed to open DuckDB store", "")
	defer store.Close()

	summary, err := store.Summarize(ctx, messages)
	utils.HandleErrorOrLogWithMessages(err, "failed to summarize messages with DuckDB", "")

	result := analyze.Build(ctx, analyze.Input{
		Messages:       messages,
		StorageSummary: summary,
		InputPath:      cfg.InputPath,
		ParserName:     parserName,
		ExtractionMode: extractionMode,
		OllamaURL:      cfg.OllamaURL,
		OllamaModel:    cfg.OllamaModel,
	})

	err = artifact.Write(ctx, cfg.OutputDir, result)
	utils.HandleErrorOrLogWithMessages(err, "failed to write artifacts", "artifacts written")

	logger.Info("build-index completed",
		"input_path", cfg.InputPath,
		"output_dir", cfg.OutputDir,
		"messages", len(messages),
		"members", len(result.Members),
		"topics", len(result.Topics),
		"inside_jokes", len(result.InsideJokes),
	)
}
