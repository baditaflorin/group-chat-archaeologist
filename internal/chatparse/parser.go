package chatparse

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/baditaflorin/group-chat-archaeologist/internal/domain"
)

const (
	AdapterWhatsAppText = "whatsapp_text"
	AdapterTelegramJSON = "telegram_json"
	AdapterSlackJSON    = "slack_json"
	AdapterDiscordCSV   = "discord_csv"
	AdapterTelegramHTML = "telegram_html"
	AdapterGenericJSON  = "json"
	AdapterPlainText    = "text"
)

var (
	whatsAppSenderLine = regexp.MustCompile(`^\[?(\d{1,4}[/-]\d{1,2}[/-]\d{1,4}),?\s+(\d{1,2}:\d{2}(?::\d{2})?\s?(?:[APap][Mm])?)\]?\s[-–]\s([^:]+?):\s(.*)$`)
	whatsAppHeaderLine = regexp.MustCompile(`^\[?(\d{1,4}[/-]\d{1,2}[/-]\d{1,4}),?\s+(\d{1,2}:\d{2}(?::\d{2})?\s?(?:[APap][Mm])?)\]?\s[-–]\s(.+)$`)
	iosSenderLine      = regexp.MustCompile(`^\[(\d{1,2}/\d{1,2}/\d{2,4}),\s+(\d{1,2}:\d{2}(?::\d{2})?)\]\s([^:]+?):\s(.*)$`)
	isoSenderLine      = regexp.MustCompile(`^\[?(\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}(?::\d{2})?)\]?\s[-–]?\s*([^:]+?):\s(.*)$`)
	tags               = regexp.MustCompile(`<[^>]+>`)
	messageStart       = regexp.MustCompile(`<div class="message[^"]*"[^>]*>`)
	fromNameBlock      = regexp.MustCompile(`(?s)<div class="from_name">\s*(.*?)\s*</div>`)
	dateTitleBlock     = regexp.MustCompile(`(?s)<div class="date" title="([^"]+)"`)
	textBlock          = regexp.MustCompile(`(?s)<div class="text">\s*(.*?)\s*</div>`)
)

type Result struct {
	Messages          []domain.Message
	Adapter           string
	AdapterConfidence float64
	Warnings          []domain.Warning
	Evidence          []string
}

type rawMessage struct {
	Timestamp string          `json:"timestamp"`
	Date      string          `json:"date"`
	Time      string          `json:"time"`
	Sender    string          `json:"sender"`
	Author    string          `json:"author"`
	From      string          `json:"from"`
	User      string          `json:"user"`
	Type      string          `json:"type"`
	Subtype   string          `json:"subtype"`
	Text      json.RawMessage `json:"text"`
	Message   string          `json:"message"`
	Content   string          `json:"content"`
	TS        string          `json:"ts"`
}

func Parse(input string) (Result, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return Result{}, fmt.Errorf("what failed: no chat messages found; why: the input is empty after normalization; now what: upload or paste a chat export")
	}

	candidates := []func(string) (Result, bool){
		parseTelegramHTML,
		parseDiscordCSV,
		parseStructuredJSON,
		parseWhatsAppText,
	}

	var best Result
	for _, candidate := range candidates {
		result, ok := candidate(trimmed)
		if !ok {
			continue
		}
		if len(result.Messages) > 0 {
			sortMessages(result.Messages)
			return result, nil
		}
		if best.Adapter == "" || result.AdapterConfidence > best.AdapterConfidence {
			best = result
		}
	}

	if best.Adapter != "" {
		return Result{}, fmt.Errorf("what failed: no messages recognized with the %s adapter; why: timestamps, senders, or message bodies were missing; now what: check the export format or regenerate it from the chat app", best.Adapter)
	}
	return Result{}, fmt.Errorf("what failed: no chat format recognized; why: the input does not resemble WhatsApp text, Telegram JSON/HTML, Slack JSON, Discord CSV, or generic JSON; now what: export the chat in one of those formats")
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

func parseStructuredJSON(input string) (Result, bool) {
	if !strings.HasPrefix(input, "[") && !strings.HasPrefix(input, "{") {
		return Result{}, false
	}

	var probe any
	if err := json.Unmarshal([]byte(input), &probe); err != nil {
		return Result{Adapter: AdapterGenericJSON, AdapterConfidence: 0.2}, false
	}

	if result, ok := parseTelegramJSON(input); ok {
		return result, true
	}
	if result, ok := parseSlackJSON(input); ok {
		return result, true
	}
	if result, ok := parseGenericJSON(input); ok {
		return result, true
	}
	return Result{Adapter: AdapterGenericJSON, AdapterConfidence: 0.35, Evidence: []string{"valid JSON, but no known message fields"}}, false
}

func parseTelegramJSON(input string) (Result, bool) {
	var envelope struct {
		Type     string       `json:"type"`
		Messages []rawMessage `json:"messages"`
	}
	if err := json.Unmarshal([]byte(input), &envelope); err != nil || len(envelope.Messages) == 0 {
		return Result{}, false
	}
	if !looksLikeTelegram(envelope.Type, envelope.Messages) {
		return Result{}, false
	}

	result := Result{
		Adapter:           AdapterTelegramJSON,
		AdapterConfidence: 0.93,
		Evidence:          []string{"JSON envelope contains messages[]", "Telegram-style date/from/text fields"},
	}
	for _, raw := range envelope.Messages {
		if raw.Type != "" && raw.Type != "message" {
			result.Warnings = append(result.Warnings, newWarning("service_message", "notice", "Skipped a Telegram service event.", "Service events describe group changes, not chat messages.", "No action needed unless you want service events included as messages.", 0, raw.Type))
			continue
		}
		sender := strings.TrimSpace(firstNonEmpty(raw.From, raw.Sender, raw.Author, raw.User))
		text := strings.TrimSpace(jsonText(raw.Text))
		stamp := strings.TrimSpace(firstNonEmpty(raw.Date, strings.TrimSpace(raw.Timestamp+" "+raw.Time)))
		if sender == "" || text == "" || stamp == "" {
			continue
		}
		parsed, err := parseTime(stamp)
		if err != nil {
			result.Warnings = append(result.Warnings, newWarning("unparsed_timestamp", "warning", "Skipped a Telegram message with an unreadable timestamp.", "The timestamp did not match the supported date formats.", "Re-export from Telegram Desktop or correct the timestamp.", 0, stamp))
			continue
		}
		result.Messages = append(result.Messages, domain.Message{
			ID:        nextID(len(result.Messages)),
			Timestamp: parsed,
			Sender:    sender,
			Text:      text,
			Source:    AdapterTelegramJSON,
		})
	}
	return result, true
}

func parseSlackJSON(input string) (Result, bool) {
	var rows []rawMessage
	if err := json.Unmarshal([]byte(input), &rows); err != nil || len(rows) == 0 {
		return Result{}, false
	}
	if firstNonEmpty(rows[0].TS, rows[0].User) == "" {
		return Result{}, false
	}

	result := Result{
		Adapter:           AdapterSlackJSON,
		AdapterConfidence: 0.91,
		Evidence:          []string{"JSON array contains Slack ts/user fields", "Slack mentions use <@USER> syntax"},
	}
	for _, raw := range rows {
		if raw.Type != "" && raw.Type != "message" {
			continue
		}
		if raw.Subtype != "" {
			result.Warnings = append(result.Warnings, newWarning("slack_subtype", "notice", "Skipped a Slack channel event.", "Slack subtype messages are usually joins, topic changes, or bot events.", "No action needed unless the event itself is part of the story.", 0, raw.Subtype))
			continue
		}
		text := strings.TrimSpace(jsonText(raw.Text))
		if strings.Contains(text, "<@") {
			result.Warnings = append(result.Warnings, newWarning("unresolved_slack_mention", "notice", "Kept a Slack user mention as an ID.", "The export did not include the user profile map needed to resolve display names.", "Add Slack user metadata in a future data run if you need names instead of IDs.", 0, text))
			text = normalizeSlackText(text)
		}
		parsed, err := parseSlackTime(raw.TS)
		if err != nil || raw.User == "" || text == "" {
			continue
		}
		result.Messages = append(result.Messages, domain.Message{
			ID:        nextID(len(result.Messages)),
			Timestamp: parsed,
			Sender:    strings.TrimSpace(raw.User),
			Text:      text,
			Source:    AdapterSlackJSON,
		})
	}
	return result, true
}

func looksLikeTelegram(exportType string, messages []rawMessage) bool {
	if strings.Contains(strings.ToLower(exportType), "group") || strings.Contains(strings.ToLower(exportType), "chat") {
		return true
	}
	for _, message := range messages {
		if message.Type != "" || message.From != "" {
			return true
		}
		if len(message.Text) > 0 && strings.HasPrefix(strings.TrimSpace(string(message.Text)), "[") {
			return true
		}
	}
	return false
}

func parseGenericJSON(input string) (Result, bool) {
	var raws []rawMessage
	if err := json.Unmarshal([]byte(input), &raws); err != nil {
		var envelope struct {
			Messages []rawMessage `json:"messages"`
		}
		if nestedErr := json.Unmarshal([]byte(input), &envelope); nestedErr != nil {
			return Result{}, false
		}
		raws = envelope.Messages
	}
	if len(raws) == 0 {
		return Result{}, false
	}

	result := Result{
		Adapter:           AdapterGenericJSON,
		AdapterConfidence: 0.75,
		Evidence:          []string{"JSON contains timestamp/sender/text-like fields"},
	}
	for _, raw := range raws {
		sender := firstNonEmpty(raw.Sender, raw.Author, raw.From, raw.User)
		text := firstNonEmpty(jsonText(raw.Text), raw.Message, raw.Content)
		stamp := strings.TrimSpace(firstNonEmpty(raw.Timestamp, raw.Date, strings.TrimSpace(raw.Date+" "+raw.Time)))
		if sender == "" || text == "" || stamp == "" {
			continue
		}
		parsed, err := parseTime(stamp)
		if err != nil {
			result.Warnings = append(result.Warnings, newWarning("unparsed_timestamp", "warning", "Skipped a JSON message with an unreadable timestamp.", "The timestamp did not match the supported date formats.", "Correct the timestamp or export with ISO-8601 dates.", 0, stamp))
			continue
		}
		result.Messages = append(result.Messages, domain.Message{
			ID:        nextID(len(result.Messages)),
			Timestamp: parsed,
			Sender:    strings.TrimSpace(sender),
			Text:      strings.TrimSpace(text),
			Source:    AdapterGenericJSON,
		})
	}
	return result, true
}

func parseDiscordCSV(input string) (Result, bool) {
	firstLine := strings.ToLower(firstLine(input))
	if !strings.Contains(firstLine, "author") || !strings.Contains(firstLine, "content") {
		return Result{}, false
	}

	reader := csv.NewReader(strings.NewReader(input))
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil || len(rows) < 2 {
		return Result{Adapter: AdapterDiscordCSV, AdapterConfidence: 0.4}, false
	}

	header := indexHeader(rows[0])
	dateIdx, okDate := header["date"]
	authorIdx, okAuthor := firstHeader(header, "author", "authorid", "sender")
	contentIdx, okContent := firstHeader(header, "content", "message", "text")
	if !okDate || !okAuthor || !okContent {
		return Result{}, false
	}

	result := Result{
		Adapter:           AdapterDiscordCSV,
		AdapterConfidence: 0.9,
		Evidence:          []string{"CSV header contains Date/Author/Content columns"},
	}
	for i, row := range rows[1:] {
		if len(row) <= max(dateIdx, authorIdx, contentIdx) {
			result.Warnings = append(result.Warnings, newWarning("short_csv_row", "warning", "Skipped a short CSV row.", "This row has fewer fields than the header.", "Check whether the CSV was truncated.", i+2, strings.Join(row, ",")))
			continue
		}
		if strings.Contains(row[contentIdx], "\n") {
			result.Warnings = append(result.Warnings, newWarning("csv_multiline_field", "notice", "Preserved a multiline Discord message.", "CSV quoting allows message text to contain line breaks.", "No action needed.", i+2, row[contentIdx]))
		}
		parsed, err := parseTime(row[dateIdx])
		if err != nil {
			result.Warnings = append(result.Warnings, newWarning("unparsed_timestamp", "warning", "Skipped a Discord row with an unreadable timestamp.", "The Date column did not match the supported date formats.", "Export with ISO-8601 dates or correct this row.", i+2, row[dateIdx]))
			continue
		}
		sender := strings.TrimSpace(row[authorIdx])
		text := strings.TrimSpace(row[contentIdx])
		if sender == "" || text == "" {
			continue
		}
		result.Messages = append(result.Messages, domain.Message{
			ID:        nextID(len(result.Messages)),
			Timestamp: parsed,
			Sender:    sender,
			Text:      text,
			Source:    AdapterDiscordCSV,
			Line:      i + 2,
		})
	}
	return result, true
}

func parseTelegramHTML(input string) (Result, bool) {
	if !strings.Contains(strings.ToLower(input), "from_name") || !strings.Contains(strings.ToLower(input), "class=\"message") {
		return Result{}, false
	}

	result := Result{
		Adapter:           AdapterTelegramHTML,
		AdapterConfidence: 0.86,
		Evidence:          []string{"HTML contains Telegram message/from_name/date/text blocks"},
	}
	blocks := messageStart.FindAllStringIndex(input, -1)
	for i, block := range blocks {
		start := block[1]
		end := len(input)
		if i+1 < len(blocks) {
			end = blocks[i+1][0]
		}
		body := input[start:end]
		sender := cleanHTML(matchFirst(fromNameBlock, body))
		stamp := cleanHTML(matchFirst(dateTitleBlock, body))
		text := cleanHTML(matchFirst(textBlock, body))
		if sender == "" || stamp == "" || text == "" {
			continue
		}
		parsed, err := parseTime(stamp)
		if err != nil {
			result.Warnings = append(result.Warnings, newWarning("unparsed_timestamp", "warning", "Skipped a Telegram HTML message with an unreadable timestamp.", "The HTML date title did not match the supported date formats.", "Re-export from Telegram Desktop or correct the date title.", 0, stamp))
			continue
		}
		result.Messages = append(result.Messages, domain.Message{
			ID:        nextID(len(result.Messages)),
			Timestamp: parsed,
			Sender:    sender,
			Text:      text,
			Source:    AdapterTelegramHTML,
		})
	}
	return result, true
}

func parseWhatsAppText(input string) (Result, bool) {
	lines := strings.Split(input, "\n")
	result := Result{
		Adapter:           AdapterWhatsAppText,
		AdapterConfidence: 0.82,
		Evidence:          []string{"line-oriented export with timestamp, sender, and message text"},
	}
	matchedHeaders := 0

	for i, line := range lines {
		lineNo := i + 1
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}

		if msg, ok, err := parseTextLine(line, len(result.Messages)+1, lineNo); err != nil {
			result.Warnings = append(result.Warnings, newWarning("unparsed_timestamp", "warning", "Skipped a message with an unreadable timestamp.", "The line looks like a message but the date/time is ambiguous.", "Check the timestamp format on this line.", lineNo, line))
			continue
		} else if ok {
			matchedHeaders++
			if strings.Contains(strings.ToLower(msg.Text), "media omitted") {
				result.Warnings = append(result.Warnings, newWarning("media_placeholder", "notice", "Found a media placeholder.", "The export references an attachment but does not include the media file.", "Re-export with media if the attachment is important.", lineNo, msg.Text))
			}
			result.Messages = append(result.Messages, msg)
			continue
		}

		if rest, ok := timestampRest(line); ok {
			if looksLikeMalformedSender(rest) {
				result.Warnings = append(result.Warnings, newWarning("malformed_timestamp_line", "warning", "Skipped a timestamp line with no message body.", "The line has a timestamp and likely a sender, but no ':' and message text.", "Check whether the export was truncated around this line.", lineNo, line))
			} else {
				result.Warnings = append(result.Warnings, newWarning("system_message", "notice", "Skipped a WhatsApp system message.", "System messages describe encryption, membership, or group settings instead of conversation text.", "No action needed unless you want service events included.", lineNo, line))
			}
			matchedHeaders++
			continue
		}

		if len(result.Messages) == 0 {
			result.Warnings = append(result.Warnings, newWarning("orphan_line", "warning", "Ignored text before the first recognized message.", "The parser has no message to attach this line to.", "Check whether the export starts midway through a message.", lineNo, line))
			continue
		}
		result.Messages[len(result.Messages)-1].Text += "\n" + strings.TrimSpace(line)
	}

	if len(result.Messages) == 0 && matchedHeaders == 0 {
		return result, false
	}
	if matchedHeaders > 0 {
		result.AdapterConfidence = min(0.97, 0.65+float64(matchedHeaders)/20)
	}
	return result, true
}

func parseTextLine(line string, next int, lineNo int) (domain.Message, bool, error) {
	if matches := isoSenderLine.FindStringSubmatch(line); len(matches) == 4 {
		parsed, err := parseTime(matches[1])
		if err != nil {
			return domain.Message{}, false, err
		}
		return message(next, parsed, matches[2], matches[3], AdapterWhatsAppText, lineNo), true, nil
	}

	if matches := iosSenderLine.FindStringSubmatch(line); len(matches) == 5 {
		parsed, err := parseTime(matches[1] + " " + matches[2])
		if err != nil {
			return domain.Message{}, false, err
		}
		return message(next, parsed, matches[3], matches[4], AdapterWhatsAppText, lineNo), true, nil
	}

	if matches := whatsAppSenderLine.FindStringSubmatch(line); len(matches) == 5 {
		parsed, err := parseTime(matches[1] + " " + matches[2])
		if err != nil {
			return domain.Message{}, false, err
		}
		return message(next, parsed, matches[3], matches[4], AdapterWhatsAppText, lineNo), true, nil
	}

	return domain.Message{}, false, nil
}

func parseTime(value string) (time.Time, error) {
	value = strings.TrimSpace(strings.ReplaceAll(value, "  ", " "))
	value = strings.TrimSuffix(value, " UTC")
	value = regexp.MustCompile(` UTC[+-]\d{2}:\d{2}$`).ReplaceAllString(value, "")
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
		"1/2/06 3:04 PM",
		"1/2/2006 3:04 PM",
		"1/2/06 15:04",
		"1/2/2006 15:04",
		"01/02/06 15:04:05",
		"01/02/2006 15:04:05",
		"02/01/06 15:04",
		"02/01/2006 15:04",
		"02/01/06 15:04:05",
		"02/01/2006 15:04:05",
		"02/01/06 3:04 PM",
		"02/01/2006 3:04 PM",
		"2-1-06 15:04",
		"2-1-2006 15:04",
		"02-01-06 15:04",
		"02-01-2006 15:04",
		"02.01.2006 15:04:05",
		"02.01.2006 15:04",
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

func parseSlackTime(value string) (time.Time, error) {
	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return time.Time{}, err
	}
	whole := int64(seconds)
	nanos := int64((seconds - float64(whole)) * 1_000_000_000)
	return time.Unix(whole, nanos).UTC(), nil
}

func jsonText(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var plain string
	if err := json.Unmarshal(raw, &plain); err == nil {
		return plain
	}
	var parts []any
	if err := json.Unmarshal(raw, &parts); err != nil {
		return ""
	}
	var out strings.Builder
	for _, part := range parts {
		switch value := part.(type) {
		case string:
			out.WriteString(value)
		case map[string]any:
			if text, ok := value["text"].(string); ok {
				out.WriteString(text)
			}
		}
	}
	return out.String()
}

func timestampRest(line string) (string, bool) {
	if matches := whatsAppHeaderLine.FindStringSubmatch(line); len(matches) == 4 {
		return strings.TrimSpace(matches[3]), true
	}
	return "", false
}

func looksLikeMalformedSender(rest string) bool {
	rest = strings.TrimSpace(rest)
	if rest == "" || strings.Contains(rest, ":") {
		return false
	}
	words := strings.Fields(rest)
	return len(words) <= 3 && regexp.MustCompile(`^[\p{L}\p{N}_ .'-]+$`).MatchString(rest)
}

func message(next int, parsed time.Time, sender, text, source string, line int) domain.Message {
	return domain.Message{
		ID:        nextID(next - 1),
		Timestamp: parsed,
		Sender:    strings.TrimSpace(sender),
		Text:      strings.TrimSpace(text),
		Source:    source,
		Line:      line,
	}
}

func firstLine(input string) string {
	if idx := strings.IndexByte(input, '\n'); idx >= 0 {
		return input[:idx]
	}
	return input
}

func indexHeader(header []string) map[string]int {
	out := map[string]int{}
	for i, name := range header {
		out[strings.ToLower(strings.TrimSpace(name))] = i
	}
	return out
}

func firstHeader(header map[string]int, names ...string) (int, bool) {
	for _, name := range names {
		if idx, ok := header[name]; ok {
			return idx, true
		}
	}
	return 0, false
}

func matchFirst(pattern *regexp.Regexp, value string) string {
	matches := pattern.FindStringSubmatch(value)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

func cleanHTML(value string) string {
	value = tags.ReplaceAllString(value, " ")
	value = html.UnescapeString(value)
	return strings.Join(strings.Fields(value), " ")
}

func normalizeSlackText(value string) string {
	value = regexp.MustCompile(`<@([^>]+)>`).ReplaceAllString(value, "$1")
	value = strings.ReplaceAll(value, "&amp;", "&")
	return strings.TrimSpace(value)
}

func newWarning(code, severity, message, why, nextStep string, line int, evidence string) domain.Warning {
	return domain.Warning{
		Code:     code,
		Severity: severity,
		Message:  message,
		Why:      why,
		NextStep: nextStep,
		Line:     line,
		Evidence: strings.TrimSpace(evidence),
	}
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
		if messages[i].Timestamp.Equal(messages[j].Timestamp) {
			return messages[i].ID < messages[j].ID
		}
		return messages[i].Timestamp.Before(messages[j].Timestamp)
	})
	for i := range messages {
		messages[i].ID = nextID(i)
	}
}

func nextID(index int) string {
	return fmt.Sprintf("m_%06d", index+1)
}

func max(values ...int) int {
	out := values[0]
	for _, value := range values[1:] {
		if value > out {
			out = value
		}
	}
	return out
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
