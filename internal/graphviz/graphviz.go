package graphviz

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/baditaflorin/group-chat-archaeologist/internal/domain"
)

func Render(ctx context.Context, outputDir string, members []domain.Member, edges []domain.IntroductionEdge) (domain.GraphArtifacts, error) {
	dotPath := filepath.Join(outputDir, "who-introduced-whom.dot")
	svgPath := filepath.Join(outputDir, "who-introduced-whom.svg")
	dot := buildDOT(members, edges)
	if err := os.WriteFile(dotPath, []byte(dot), 0o644); err != nil {
		return domain.GraphArtifacts{}, fmt.Errorf("write dot: %w", err)
	}

	result := domain.GraphArtifacts{
		DOTPath:  "data/v1/who-introduced-whom.dot",
		SVGPath:  "data/v1/who-introduced-whom.svg",
		Renderer: "graphviz-dot",
	}

	if err := exec.CommandContext(ctx, "dot", "-Tsvg", dotPath, "-o", svgPath).Run(); err != nil {
		result.RenderError = err.Error()
		if fallbackErr := os.WriteFile(svgPath, []byte(fallbackSVG()), 0o644); fallbackErr != nil {
			return result, fmt.Errorf("render graphviz: %w; write fallback: %w", err, fallbackErr)
		}
		return result, nil
	}

	result.Rendered = true
	return result, nil
}

func buildDOT(members []domain.Member, edges []domain.IntroductionEdge) string {
	var b strings.Builder
	b.WriteString("digraph introductions {\n")
	b.WriteString("  graph [rankdir=LR, bgcolor=\"transparent\", pad=\"0.35\"];\n")
	b.WriteString("  node [shape=box, style=\"rounded,filled\", fillcolor=\"#f4f0e6\", color=\"#2f5f53\", fontname=\"Inter\", fontsize=12];\n")
	b.WriteString("  edge [color=\"#8a5a44\", fontname=\"Inter\", fontsize=10, arrowsize=0.7];\n")

	names := make([]string, 0, len(members))
	for _, member := range members {
		names = append(names, member.Name)
	}
	sort.Strings(names)
	for _, name := range names {
		b.WriteString("  " + quote(name) + ";\n")
	}
	for _, edge := range edges {
		b.WriteString("  " + quote(edge.From) + " -> " + quote(edge.To) + " [label=" + quote(edge.FirstMentionAt[:10]) + "];\n")
	}
	b.WriteString("}\n")
	return b.String()
}

func quote(value string) string {
	escaped := strings.ReplaceAll(value, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	return "\"" + escaped + "\""
}

func fallbackSVG() string {
	return `<svg xmlns="http://www.w3.org/2000/svg" width="720" height="240" viewBox="0 0 720 240"><rect width="720" height="240" fill="#f4f0e6"/><text x="32" y="120" font-family="Inter, sans-serif" font-size="20" fill="#16251f">GraphViz rendering unavailable. DOT artifact is available.</text></svg>`
}
