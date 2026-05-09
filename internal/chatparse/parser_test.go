package chatparse

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/baditaflorin/group-chat-archaeologist/internal/domain"
	"github.com/baditaflorin/group-chat-archaeologist/internal/extract"
	"github.com/stretchr/testify/require"
)

func TestParseWhatsAppStyle(t *testing.T) {
	input := `1/2/21, 9:05 AM - Ana: I invited Bogdan for the archive map.
This continuation stays with Ana.
1/2/21, 9:07 AM - Bogdan: Hello from the old chat.`

	result, err := Parse(input)
	require.NoError(t, err)
	require.Equal(t, AdapterWhatsAppText, result.Adapter)
	require.Len(t, result.Messages, 2)
	require.Equal(t, "Ana", result.Messages[0].Sender)
	require.Contains(t, result.Messages[0].Text, "continuation")
	require.Equal(t, "Bogdan", result.Messages[1].Sender)
}

func TestParseJSONEnvelope(t *testing.T) {
	input := `{"messages":[{"timestamp":"2022-03-04T10:00:00Z","sender":"Carmen","text":"sarmale index"}]}`

	result, err := Parse(input)
	require.NoError(t, err)
	require.Equal(t, AdapterGenericJSON, result.Adapter)
	require.Len(t, result.Messages, 1)
	require.Equal(t, "Carmen", result.Messages[0].Sender)
}

func TestRealDataFixtures(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "test", "fixtures", "realdata")
	inputs, err := filepath.Glob(filepath.Join(fixtureDir, "*.*"))
	require.NoError(t, err)

	for _, inputPath := range inputs {
		if strings.HasSuffix(inputPath, ".expected.json") {
			continue
		}
		name := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
		t.Run(name, func(t *testing.T) {
			expected := readExpected(t, strings.TrimSuffix(inputPath, filepath.Ext(inputPath))+".expected.json")
			extracted, err := extract.Text(context.Background(), inputPath, "")
			require.NoError(t, err)
			result, err := Parse(extracted.Text)
			require.NoError(t, err)
			warnings := append(extracted.Warnings, result.Warnings...)

			require.Equal(t, expected.Adapter, result.Adapter)
			require.GreaterOrEqual(t, len(result.Messages), expected.MinMessages)
			require.Subset(t, members(result.Messages), expected.Members)
			require.Subset(t, warningCodes(warnings), expected.RequiresWarningCodes)
			for _, forbidden := range expected.ForbiddenSnippetInMessageText {
				for _, message := range result.Messages {
					require.NotContains(t, message.Text, forbidden)
				}
			}
		})
	}
}

type expectedFixture struct {
	Adapter                       string   `json:"adapter"`
	MinMessages                   int      `json:"minMessages"`
	Members                       []string `json:"members"`
	RequiresWarningCodes          []string `json:"requiresWarningCodes"`
	ForbiddenSnippetInMessageText []string `json:"forbiddenSnippetInMessageText"`
}

func readExpected(t *testing.T, path string) expectedFixture {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var expected expectedFixture
	require.NoError(t, json.Unmarshal(data, &expected))
	return expected
}

func members(messages []domain.Message) []string {
	seen := map[string]bool{}
	for _, message := range messages {
		seen[message.Sender] = true
	}
	out := make([]string, 0, len(seen))
	for member := range seen {
		out = append(out, member)
	}
	sort.Strings(out)
	return out
}

func warningCodes(warnings []domain.Warning) []string {
	seen := map[string]bool{}
	for _, warning := range warnings {
		seen[warning.Code] = true
	}
	out := make([]string, 0, len(seen))
	for code := range seen {
		out = append(out, code)
	}
	sort.Strings(out)
	return out
}
