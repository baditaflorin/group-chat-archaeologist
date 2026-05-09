package analyze

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/baditaflorin/group-chat-archaeologist/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestBuildFindsIntroductionsAndJokes(t *testing.T) {
	messages := []domain.Message{
		{ID: "m_000001", Timestamp: mustTime("2021-01-01T09:00:00Z"), Sender: "Ana", Text: "I invited Bogdan to track the sarmale index."},
		{ID: "m_000002", Timestamp: mustTime("2021-01-01T09:10:00Z"), Sender: "Bogdan", Text: "The sarmale index is already scientific."},
		{ID: "m_000003", Timestamp: mustTime("2021-02-01T09:10:00Z"), Sender: "Carmen", Text: "Sarmale index forever."},
	}
	summary := domain.StorageSummary{MemberStats: []domain.MemberStat{
		{Name: "Ana", MessageCount: 1, FirstMessageAt: messages[0].Timestamp, LastMessageAt: messages[0].Timestamp},
		{Name: "Bogdan", MessageCount: 1, FirstMessageAt: messages[1].Timestamp, LastMessageAt: messages[1].Timestamp},
		{Name: "Carmen", MessageCount: 1, FirstMessageAt: messages[2].Timestamp, LastMessageAt: messages[2].Timestamp},
	}}

	dashboard := Build(context.Background(), Input{Messages: messages, StorageSummary: summary, InputPath: "sample.txt", ParserName: "text", ExtractionMode: "txt"})

	require.NotEmpty(t, dashboard.Topics)
	require.Len(t, dashboard.Introductions, 1)
	require.Equal(t, "Bogdan", dashboard.Introductions[0].To)
	require.NotEmpty(t, dashboard.InsideJokes)
}

func TestBuildIsDeterministicWithFixedGeneratedAt(t *testing.T) {
	messages := []domain.Message{
		{ID: "m_000001", Timestamp: mustTime("2024-01-01T09:00:00Z"), Sender: "Ana", Text: "I invited Bogdan for the train plan."},
		{ID: "m_000002", Timestamp: mustTime("2024-01-01T09:05:00Z"), Sender: "Bogdan", Text: "Train plan accepted."},
	}
	summary := domain.StorageSummary{MemberStats: []domain.MemberStat{
		{Name: "Ana", MessageCount: 1, FirstMessageAt: messages[0].Timestamp, LastMessageAt: messages[0].Timestamp},
		{Name: "Bogdan", MessageCount: 1, FirstMessageAt: messages[1].Timestamp, LastMessageAt: messages[1].Timestamp},
	}}
	input := Input{
		Messages:          messages,
		StorageSummary:    summary,
		InputPath:         "sample.txt",
		ParserName:        "whatsapp_text",
		Adapter:           "whatsapp_text",
		AdapterConfidence: 0.9,
		ExtractionMode:    "txt",
		GeneratedAt:       "2026-05-09T00:00:00Z",
		Concurrency:       4,
		SaveEvery:         5000,
	}

	first, err := json.Marshal(Build(context.Background(), input))
	require.NoError(t, err)
	second, err := json.Marshal(Build(context.Background(), input))
	require.NoError(t, err)
	require.Equal(t, string(first), string(second))
}

func mustTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return parsed
}
