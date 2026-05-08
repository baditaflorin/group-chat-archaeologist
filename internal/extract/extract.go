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
)

func Text(ctx context.Context, inputPath, tikaURL string) (string, string, error) {
	ext := strings.ToLower(filepath.Ext(inputPath))
	if isTextLike(ext) {
		data, err := os.ReadFile(inputPath)
		if err != nil {
			return "", "", fmt.Errorf("read input: %w", err)
		}
		return string(data), strings.TrimPrefix(ext, "."), nil
	}

	if tikaURL == "" {
		return "", "", fmt.Errorf("input %q needs Tika extraction; set --tika_url or TIKA_SERVER_URL", inputPath)
	}

	data, err := os.ReadFile(inputPath)
	if err != nil {
		return "", "", fmt.Errorf("read binary input: %w", err)
	}

	text, err := tika(ctx, tikaURL, data)
	if err != nil {
		return "", "", err
	}
	return text, "tika", nil
}

func isTextLike(ext string) bool {
	switch ext {
	case ".txt", ".json", ".csv", ".tsv", ".md", ".log":
		return true
	default:
		return false
	}
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
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read Tika response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("Tika returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return string(body), nil
}
