package chatparse

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseWhatsAppStyle(t *testing.T) {
	input := `1/2/21, 9:05 AM - Ana: I invited Bogdan for the archive map.
This continuation stays with Ana.
1/2/21, 9:07 AM - Bogdan: Hello from the old chat.`

	messages, parser, err := Parse(input)
	require.NoError(t, err)
	require.Equal(t, "text", parser)
	require.Len(t, messages, 2)
	require.Equal(t, "Ana", messages[0].Sender)
	require.Contains(t, messages[0].Text, "continuation")
	require.Equal(t, "Bogdan", messages[1].Sender)
}

func TestParseJSONEnvelope(t *testing.T) {
	input := `{"messages":[{"timestamp":"2022-03-04T10:00:00Z","sender":"Carmen","text":"sarmale index"}]}`

	messages, parser, err := Parse(input)
	require.NoError(t, err)
	require.Equal(t, "json", parser)
	require.Len(t, messages, 1)
	require.Equal(t, "Carmen", messages[0].Sender)
}
