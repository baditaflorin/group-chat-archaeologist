package chatparse

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/baditaflorin/group-chat-archaeologist/internal/domain"
)

var (
	whatsAppLine = regexp.MustCompile(`^\[?(\d{1,2}/\d{1,2}/\d{2,4}),?\s+(\d{1,2}:\d{2}(?::\d{2})?\s?(?:[APap][Mm])?)\]?\s[-–]\s([^:]+?):\s(.*)$`)
	isoLine      = regexp.MustCompile(`^\[?(\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}(?::\d{2})?)\]?\s[-–]?\s*([^:]+?):\s(.*)$`)
)

type rawMessage struct {
	Timestamp string `json:"timestamp"`
	Date      string `json:"date"`
	Time      string `json:"time"`
	Sender    string `json:"sender"`
	Author    string `json:"author"`
	From      string `json:"from"`
	Text      string `json:"text"`
	Message   string `json:"message"`
	Content   string `json:"content"`
}

func Parse(input string) ([]domain.Message, string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, "", fmt.Errorf("empty chat export")
	}

	if strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{") {
		messages, err := parseJSON(trimmed)
		if err == nil && len(messages) > 0 {
			return messages, "json", nil
		}
	}

	messages, err := parsePlainText(trimmed)
	if err != nil {
		return nil, "", err
	}
	return messages, "text", nil
}

func Filter(messages []domain.Message, start, end time.Time) []domain.Message {
	filtered := make([]domain.Message, 0, len(messages))
	for _, msg := range messages {
		if !start.IsZero() && msg.Timestamp.Before(start) {
			continue
		}
		if !end.IsZero() && msg.Timestamp.After(end) {
			continue
		}
		filtered = append(filtered, msg)
	}
	return filtered
}

func parseJSON(input string) ([]domain.Message, error) {
	var raws []rawMessage
	if err := json.Unmarshal([]byte(input), &raws); err != nil {
		var envelope struct {
			Messages []rawMessage `json:"messages"`
		}
		if nestedErr := json.Unmarshal([]byte(input), &envelope); nestedErr != nil {
			return nil, err
		}
		raws = envelope.Messages
	}

	messages := make([]domain.Message, 0, len(raws))
	for _, raw := range raws {
		sender := firstNonEmpty(raw.Sender, raw.Author, raw.From)
		text := firstNonEmpty(raw.Text, raw.Message, raw.Content)
		stamp := strings.TrimSpace(firstNonEmpty(raw.Timestamp, strings.TrimSpace(raw.Date+" "+raw.Time)))
		if sender == "" || text == "" || stamp == "" {
			continue
		}
		parsed, err := parseTime(stamp)
		if err != nil {
			return nil, fmt.Errorf("parse JSON timestamp %q: %w", stamp, err)
		}
		messages = append(messages, domain.Message{
			ID:        fmt.Sprintf("m_%06d", len(messages)+1),
			Timestamp: parsed,
			Sender:    strings.TrimSpace(sender),
			Text:      strings.TrimSpace(text),
		})
	}

	sortMessages(messages)
	return messages, nil
}

func parsePlainText(input string) ([]domain.Message, error) {
	lines := strings.Split(input, "\n")
	messages := make([]domain.Message, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}

		if msg, ok, err := parseLine(line, len(messages)+1); err != nil {
			return nil, err
		} else if ok {
			messages = append(messages, msg)
			continue
		}

		if len(messages) == 0 {
			continue
		}
		messages[len(messages)-1].Text += "\n" + strings.TrimSpace(line)
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages recognized")
	}

	sortMessages(messages)
	return messages, nil
}

func parseLine(line string, nextID int) (domain.Message, bool, error) {
	if matches := isoLine.FindStringSubmatch(line); len(matches) == 4 {
		parsed, err := parseTime(matches[1])
		if err != nil {
			return domain.Message{}, false, err
		}
		return domain.Message{
			ID:        fmt.Sprintf("m_%06d", nextID),
			Timestamp: parsed,
			Sender:    strings.TrimSpace(matches[2]),
			Text:      strings.TrimSpace(matches[3]),
		}, true, nil
	}

	if matches := whatsAppLine.FindStringSubmatch(line); len(matches) == 5 {
		parsed, err := parseTime(matches[1] + " " + matches[2])
		if err != nil {
			return domain.Message{}, false, err
		}
		return domain.Message{
			ID:        fmt.Sprintf("m_%06d", nextID),
			Timestamp: parsed,
			Sender:    strings.TrimSpace(matches[3]),
			Text:      strings.TrimSpace(matches[4]),
		}, true, nil
	}

	return domain.Message{}, false, nil
}

func parseTime(value string) (time.Time, error) {
	value = strings.TrimSpace(strings.ReplaceAll(value, "  ", " "))
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"1/2/06 3:04 PM",
		"1/2/2006 3:04 PM",
		"1/2/06 15:04",
		"1/2/2006 15:04",
		"02/01/06 15:04",
		"02/01/2006 15:04",
		"02/01/06 3:04 PM",
		"02/01/2006 3:04 PM",
	}

	var lastErr error
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return parsed.UTC(), nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func sortMessages(messages []domain.Message) {
	sort.SliceStable(messages, func(i, j int) bool {
		return messages[i].Timestamp.Before(messages[j].Timestamp)
	})
	for i := range messages {
		messages[i].ID = fmt.Sprintf("m_%06d", i+1)
	}
}
