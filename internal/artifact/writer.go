package artifact

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/baditaflorin/group-chat-archaeologist/internal/domain"
	"github.com/baditaflorin/group-chat-archaeologist/internal/graphviz"
)

type Meta struct {
	GeneratedAt     string `json:"generatedAt"`
	SchemaVersion   string `json:"schemaVersion"`
	SourceCommit    string `json:"sourceCommit"`
	InputSHA256     string `json:"inputSha256"`
	MessageCount    int    `json:"messageCount"`
	WarningCount    int    `json:"warningCount"`
	AppVersion      string `json:"appVersion"`
	GraphRendered   bool   `json:"graphRendered"`
	ArtifactVersion string `json:"artifactVersion"`
}

func Write(ctx context.Context, outputDir string, dashboard domain.Dashboard) error {
	tmpDir := outputDir + ".tmp"
	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("remove tmp dir: %w", err)
	}
	if err := os.MkdirAll(tmpDir, 0o750); err != nil {
		return fmt.Errorf("create tmp dir: %w", err)
	}

	graph, err := graphviz.Render(ctx, tmpDir, dashboard.Members, dashboard.Introductions)
	if err != nil {
		return err
	}
	dashboard.Graph = graph

	if err := writeJSON(filepath.Join(tmpDir, "chat-archaeology.json"), dashboard); err != nil {
		return err
	}
	meta := Meta{
		GeneratedAt:     dashboard.GeneratedAt,
		SchemaVersion:   dashboard.SchemaVersion,
		SourceCommit:    dashboard.Source.SourceCommit,
		InputSHA256:     dashboard.Source.InputSHA256,
		MessageCount:    dashboard.Source.MessageCount,
		WarningCount:    dashboard.Source.WarningCount,
		AppVersion:      dashboard.Source.AppVersion,
		GraphRendered:   dashboard.Graph.Rendered,
		ArtifactVersion: "1",
	}
	if err := writeJSON(filepath.Join(tmpDir, "chat-archaeology.meta.json"), meta); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputDir), 0o750); err != nil {
		return fmt.Errorf("create output parent: %w", err)
	}
	if err := os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("remove old output: %w", err)
	}
	if err := os.Rename(tmpDir, outputDir); err != nil {
		return fmt.Errorf("promote artifacts: %w", err)
	}
	return nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", filepath.Base(path), err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", filepath.Base(path), err)
	}
	return nil
}
