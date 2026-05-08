package analyze

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/baditaflorin/group-chat-archaeologist/internal/domain"
)

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

func enrichTopicsWithOllama(ctx context.Context, baseURL, model string, topics []domain.TopicPeriod) ([]domain.TopicPeriod, bool) {
	if model == "" {
		model = "llama3.2"
	}

	client := &http.Client{Timeout: 4 * time.Second}
	enriched := append([]domain.TopicPeriod(nil), topics...)
	used := false
	for i := range enriched {
		prompt := "Create a short, neutral topic label of at most 5 words for these chat keywords: " + strings.Join(enriched[i].Keywords, ", ")
		body, err := json.Marshal(ollamaRequest{Model: model, Prompt: prompt, Stream: false})
		if err != nil {
			return topics, used
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/api/generate", bytes.NewReader(body))
		if err != nil {
			return topics, used
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return topics, used
		}
		var parsed ollamaResponse
		err = json.NewDecoder(resp.Body).Decode(&parsed)
		_ = resp.Body.Close()
		if err != nil || resp.StatusCode >= http.StatusBadRequest {
			return topics, used
		}
		label := sanitizeLabel(parsed.Response)
		if label != "" {
			enriched[i].Label = label
			used = true
		}
	}
	return enriched, used
}

func sanitizeLabel(label string) string {
	label = strings.Trim(label, " \n\t\"'`.")
	label = strings.Join(strings.Fields(label), " ")
	if len(label) > 60 {
		label = label[:60]
	}
	return label
}
