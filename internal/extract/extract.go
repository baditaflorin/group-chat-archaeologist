package extract

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/baditaflorin/group-chat-archaeologist/internal/domain"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type Result struct {
	Text               string
	ExtractionMode     string
	Warnings           []domain.Warning
	NormalizationSteps []string
}

func Text(ctx context.Context, inputPath, tikaURL string) (Result, error) {
	ext := strings.ToLower(filepath.Ext(inputPath))
	if isTextLike(ext) {
		data, err := os.ReadFile(inputPath)
		if err != nil {
			return Result{}, fmt.Errorf("read input: %w", err)
		}
		text, warnings, steps := normalizeBytes(data)
		return Result{
			Text:               text,
			ExtractionMode:     strings.TrimPrefix(ext, "."),
			Warnings:           warnings,
			NormalizationSteps: steps,
		}, nil
	}

	if tikaURL == "" {
		return Result{}, fmt.Errorf("input %q needs Tika extraction; set --tika_url or TIKA_SERVER_URL", inputPath)
	}

	data, err := os.ReadFile(inputPath)
	if err != nil {
		return Result{}, fmt.Errorf("read binary input: %w", err)
	}

	text, err := tika(ctx, tikaURL, data)
	if err != nil {
		return Result{}, err
	}
	normalized, warnings, steps := normalizeBytes([]byte(text))
	return Result{
		Text:               normalized,
		ExtractionMode:     "tika",
		Warnings:           warnings,
		NormalizationSteps: steps,
	}, nil
}

func isTextLike(ext string) bool {
	switch ext {
	case ".txt", ".json", ".csv", ".tsv", ".md", ".log", ".html", ".htm":
		return true
	default:
		return false
	}
}

func normalizeBytes(data []byte) (string, []domain.Warning, []string) {
	warnings := []domain.Warning{}
	steps := []string{}

	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
		warnings = append(warnings, warning("normalized_bom", "notice", "Removed a UTF-8 byte-order mark from the export.", "Some chat apps add an invisible marker before the first timestamp.", "No action needed."))
		steps = append(steps, "removed UTF-8 BOM")
	}

	var text string
	if utf8.Valid(data) {
		text = string(data)
		steps = append(steps, "decoded UTF-8")
	} else {
		decoded, _, err := transform.String(charmap.Windows1252.NewDecoder(), string(data))
		if err == nil {
			text = decoded
			warnings = append(warnings, warning("normalized_encoding", "notice", "Decoded the export as Windows-1252.", "The input was not valid UTF-8, which is common in older desktop exports.", "Verify names and punctuation if they look unusual."))
			steps = append(steps, "decoded Windows-1252")
		} else {
			text = strings.ToValidUTF8(string(data), "�")
			warnings = append(warnings, warning("invalid_utf8_replaced", "warning", "Replaced invalid text bytes during import.", "The export contains byte sequences that are not valid UTF-8 or Windows-1252.", "Re-export the chat as UTF-8 if names or messages look corrupted."))
			steps = append(steps, "replaced invalid UTF-8")
		}
	}

	if strings.Contains(text, "\r\n") {
		text = strings.ReplaceAll(text, "\r\n", "\n")
		warnings = append(warnings, warning("normalized_crlf", "notice", "Normalized Windows line endings.", "CRLF line endings can hide multiline message boundaries.", "No action needed."))
		steps = append(steps, "normalized CRLF")
	}
	if strings.Contains(text, "\r") {
		text = strings.ReplaceAll(text, "\r", "\n")
		steps = append(steps, "normalized CR")
	}
	if strings.Contains(text, "\u00a0") {
		text = strings.ReplaceAll(text, "\u00a0", " ")
		warnings = append(warnings, warning("normalized_nbsp", "notice", "Normalized non-breaking spaces.", "Some exports use invisible spacing characters inside timestamps and names.", "No action needed."))
		steps = append(steps, "normalized NBSP")
	}

	replacer := strings.NewReplacer("\u200e", "", "\u200f", "", "\u202a", "", "\u202b", "", "\u202c", "", "\u202d", "", "\u202e", "")
	clean := replacer.Replace(text)
	if clean != text {
		text = clean
		warnings = append(warnings, warning("normalized_direction_marks", "notice", "Removed invisible direction marks.", "RTL/LTR markers can prevent timestamp recognition.", "No action needed."))
		steps = append(steps, "removed direction marks")
	}

	if len(data) > 1024*1024 || strings.Count(text, "\n") > 10000 {
		warnings = append(warnings, warning("large_input", "notice", "Large chat export detected.", "The archive is big enough that analysis can take noticeably longer.", "Let the build finish; if the browser feels slow, regenerate data offline and publish the artifacts."))
	}

	return text, warnings, uniqueStrings(steps)
}

func warning(code, severity, message, why, nextStep string) domain.Warning {
	return domain.Warning{Code: code, Severity: severity, Message: message, Why: why, NextStep: nextStep}
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func tika(ctx context.Context, baseURL string, data []byte) (string, error) {
	url := strings.TrimRight(baseURL, "/") + "/tika"
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create Tika request: %w", err)
	}
	req.Header.Set("Accept", "text/plain")
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call Tika: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read Tika response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("tika returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return string(body), nil
}
