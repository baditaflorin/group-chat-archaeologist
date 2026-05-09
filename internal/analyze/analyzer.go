package analyze

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/baditaflorin/group-chat-archaeologist/internal/domain"
)

const (
	repositoryURL = "https://github.com/baditaflorin/group-chat-archaeologist"
	payPalURL     = "https://www.paypal.com/paypalme/florinbadita"
	appVersion    = "0.2.0"
)

type Input struct {
	Messages           []domain.Message
	StorageSummary     domain.StorageSummary
	InputPath          string
	ParserName         string
	Adapter            string
	AdapterConfidence  float64
	AdapterEvidence    []string
	ExtractionMode     string
	NormalizationSteps []string
	Warnings           []domain.Warning
	GeneratedAt        string
	Start              time.Time
	End                time.Time
	Concurrency        int
	SaveEvery          int
	OllamaURL          string
	OllamaModel        string
}

func Build(ctx context.Context, input Input) domain.Dashboard {
	messages := append([]domain.Message(nil), input.Messages...)
	sort.SliceStable(messages, func(i, j int) bool {
		return messages[i].Timestamp.Before(messages[j].Timestamp)
	})

	topics := topicTimeline(messages)
	llmUsed := false
	if input.OllamaURL != "" {
		if enriched, ok := enrichTopicsWithOllama(ctx, input.OllamaURL, input.OllamaModel, topics); ok {
			topics = enriched
			llmUsed = true
		}
	}
	generatedAt := input.GeneratedAt
	if generatedAt == "" {
		generatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	warnings := normalizeWarnings(input.Warnings)

	return domain.Dashboard{
		SchemaVersion:   domain.SchemaVersion,
		GeneratedAt:     generatedAt,
		RepositoryURL:   repositoryURL,
		PayPalURL:       payPalURL,
		Source:          sourceSummary(input, messages, llmUsed, warnings),
		Members:         members(input.StorageSummary),
		Topics:          topics,
		Introductions:   introductionEdges(messages),
		InsideJokes:     insideJokes(messages),
		Departures:      departures(messages),
		NotableMessages: notableMessages(messages),
		Warnings:        warnings,
		Debug: domain.DebugInfo{
			AdapterEvidence: cloneStrings(input.AdapterEvidence),
			ParseWarnings:   len(warnings),
		},
	}
}

func sourceSummary(input Input, messages []domain.Message, llmUsed bool, warnings []domain.Warning) domain.SourceSummary {
	var first, last string
	if len(messages) > 0 {
		first = messages[0].Timestamp.Format(time.RFC3339)
		last = messages[len(messages)-1].Timestamp.Format(time.RFC3339)
	}
	adapter := input.Adapter
	if adapter == "" {
		adapter = input.ParserName
	}

	return domain.SourceSummary{
		InputName:          filepath.Base(input.InputPath),
		InputSHA256:        fileSHA256(input.InputPath),
		Parser:             input.ParserName,
		Adapter:            adapter,
		AdapterConfidence:  roundConfidence(input.AdapterConfidence),
		ExtractionMode:     input.ExtractionMode,
		NormalizationSteps: cloneStrings(input.NormalizationSteps),
		AnalyticsEngine:    input.StorageSummary.Engine,
		MessageCount:       len(messages),
		MemberCount:        len(input.StorageSummary.MemberStats),
		FirstMessageAt:     first,
		LastMessageAt:      last,
		WarningCount:       len(warnings),
		LLMProvider:        providerName(input.OllamaURL),
		LLMModel:           input.OllamaModel,
		LLMUsed:            llmUsed,
		SourceCommit:       gitCommit(),
		AppVersion:         appVersion,
		Parameters:         parameters(input),
	}
}

func members(summary domain.StorageSummary) []domain.Member {
	out := make([]domain.Member, 0, len(summary.MemberStats))
	for _, stat := range summary.MemberStats {
		out = append(out, domain.Member{
			Name:           stat.Name,
			MessageCount:   stat.MessageCount,
			FirstMessageAt: stat.FirstMessageAt.Format(time.RFC3339),
			LastMessageAt:  stat.LastMessageAt.Format(time.RFC3339),
		})
	}
	return out
}

func topicTimeline(messages []domain.Message) []domain.TopicPeriod {
	type bucket struct {
		start   time.Time
		end     time.Time
		msgs    []domain.Message
		words   map[string]int
		members map[string]int
	}

	buckets := map[string]*bucket{}
	for _, msg := range messages {
		key := msg.Timestamp.Format("2006-01")
		if _, ok := buckets[key]; !ok {
			start := time.Date(msg.Timestamp.Year(), msg.Timestamp.Month(), 1, 0, 0, 0, 0, time.UTC)
			buckets[key] = &bucket{
				start:   start,
				end:     start.AddDate(0, 1, -1),
				words:   map[string]int{},
				members: map[string]int{},
			}
		}
		b := buckets[key]
		b.msgs = append(b.msgs, msg)
		b.members[msg.Sender]++
		for _, token := range tokenize(msg.Text) {
			if !stopWords[token] {
				b.words[token]++
			}
		}
	}

	keys := make([]string, 0, len(buckets))
	for key := range buckets {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	topics := make([]domain.TopicPeriod, 0, len(keys))
	for _, key := range keys {
		b := buckets[key]
		keywords := topKeys(b.words, 5)
		topMembers := topKeys(b.members, 3)
		label := "General memory"
		if len(keywords) > 0 {
			label = titleFromKeywords(keywords)
		}
		topics = append(topics, domain.TopicPeriod{
			ID:           "topic_" + strings.ReplaceAll(key, "-", "_"),
			Label:        label,
			Start:        b.start.Format("2006-01-02"),
			End:          b.end.Format("2006-01-02"),
			MessageCount: len(b.msgs),
			Keywords:     keywords,
			TopMembers:   topMembers,
			Summary:      fmt.Sprintf("%d messages led by %s.", len(b.msgs), strings.Join(topMembers, ", ")),
			Confidence: confidence(0.55+minFloat(0.35, float64(len(b.msgs))/40), []string{
				fmt.Sprintf("month bucket %s contains %d parsed messages", key, len(b.msgs)),
				fmt.Sprintf("top keywords: %s", strings.Join(keywords, ", ")),
			}),
		})
	}
	return topics
}

func introductionEdges(messages []domain.Message) []domain.IntroductionEdge {
	firstSent := map[string]time.Time{}
	for _, msg := range messages {
		if _, ok := firstSent[msg.Sender]; !ok {
			firstSent[msg.Sender] = msg.Timestamp
		}
	}

	members := make([]string, 0, len(firstSent))
	for member := range firstSent {
		members = append(members, member)
	}
	sort.Strings(members)

	seen := map[string]bool{}
	edges := []domain.IntroductionEdge{}
	for _, msg := range messages {
		body := strings.ToLower(msg.Text)
		for _, target := range members {
			if target == msg.Sender || seen[target] || !msg.Timestamp.Before(firstSent[target]) {
				continue
			}
			if mentionsName(body, target) {
				seen[target] = true
				edges = append(edges, domain.IntroductionEdge{
					From:           msg.Sender,
					To:             target,
					FirstMentionAt: msg.Timestamp.Format(time.RFC3339),
					MessageID:      msg.ID,
					Snippet:        snippet(msg.Text, 140),
					Confidence: confidence(0.78, []string{
						fmt.Sprintf("%s was mentioned before their first message", target),
						fmt.Sprintf("first %s message was at %s", target, firstSent[target].Format(time.RFC3339)),
					}),
				})
			}
		}
	}
	return edges
}

func insideJokes(messages []domain.Message) []domain.InsideJoke {
	type phraseStat struct {
		count        int
		origin       domain.Message
		participants map[string]bool
	}

	stats := map[string]*phraseStat{}
	for _, msg := range messages {
		tokens := tokenize(msg.Text)
		seenInMessage := map[string]bool{}
		for n := 2; n <= 4; n++ {
			for i := 0; i+n <= len(tokens); i++ {
				phraseTokens := tokens[i : i+n]
				if tooBoring(phraseTokens) {
					continue
				}
				phrase := strings.Join(phraseTokens, " ")
				if seenInMessage[phrase] {
					continue
				}
				seenInMessage[phrase] = true
				if _, ok := stats[phrase]; !ok {
					stats[phrase] = &phraseStat{origin: msg, participants: map[string]bool{}}
				}
				stats[phrase].count++
				stats[phrase].participants[msg.Sender] = true
			}
		}
	}

	jokes := []domain.InsideJoke{}
	for phrase, stat := range stats {
		if stat.count < 2 || len(stat.participants) < 2 {
			continue
		}
		participants := make([]string, 0, len(stat.participants))
		for participant := range stat.participants {
			participants = append(participants, participant)
		}
		sort.Strings(participants)
		jokes = append(jokes, domain.InsideJoke{
			Phrase:       phrase,
			OriginAt:     stat.origin.Timestamp.Format(time.RFC3339),
			OriginSender: stat.origin.Sender,
			OriginID:     stat.origin.ID,
			Occurrences:  stat.count,
			Participants: participants,
			Snippet:      snippet(stat.origin.Text, 160),
			Confidence: confidence(0.45+minFloat(0.45, float64(stat.count+len(participants))/12), []string{
				fmt.Sprintf("phrase repeated %d times", stat.count),
				fmt.Sprintf("used by %d participants", len(participants)),
			}),
		})
	}

	jokes = dedupeInsideJokes(jokes)
	sort.SliceStable(jokes, func(i, j int) bool {
		if jokes[i].Occurrences == jokes[j].Occurrences {
			return jokes[i].OriginAt < jokes[j].OriginAt
		}
		return jokes[i].Occurrences > jokes[j].Occurrences
	})
	if len(jokes) > 8 {
		return jokes[:8]
	}
	return jokes
}

func departures(messages []domain.Message) []domain.Departure {
	if len(messages) == 0 {
		return nil
	}

	type activity struct {
		first domain.Message
		last  domain.Message
		count int
	}
	byMember := map[string]activity{}
	for _, msg := range messages {
		current, ok := byMember[msg.Sender]
		if !ok {
			current.first = msg
		}
		current.last = msg
		current.count++
		byMember[msg.Sender] = current
	}

	final := messages[len(messages)-1].Timestamp
	out := make([]domain.Departure, 0, len(byMember))
	for member, active := range byMember {
		gap := int(final.Sub(active.last.Timestamp).Hours() / 24)
		span := int(active.last.Timestamp.Sub(active.first.Timestamp).Hours() / 24)
		status, interpretation := departureStatus(gap)
		out = append(out, domain.Departure{
			Member:          member,
			Status:          status,
			LastMessageAt:   active.last.Timestamp.Format(time.RFC3339),
			DaysSinceActive: gap,
			ActiveSpanDays:  span,
			LastSnippet:     snippet(active.last.Text, 140),
			Interpretation:  interpretation,
			Confidence: confidence(departureConfidence(gap, span, active.count), []string{
				fmt.Sprintf("last message is %d days before archive end", gap),
				fmt.Sprintf("%d messages across %d active days", active.count, span),
			}),
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].DaysSinceActive == out[j].DaysSinceActive {
			return out[i].Member < out[j].Member
		}
		return out[i].DaysSinceActive > out[j].DaysSinceActive
	})
	return out
}

func notableMessages(messages []domain.Message) []domain.NotableMessage {
	firstByMember := map[string]bool{}
	out := []domain.NotableMessage{}
	for _, msg := range messages {
		if !firstByMember[msg.Sender] {
			firstByMember[msg.Sender] = true
			out = append(out, domain.NotableMessage{
				ID:      msg.ID,
				At:      msg.Timestamp.Format(time.RFC3339),
				Sender:  msg.Sender,
				Kind:    "first-message",
				Snippet: snippet(msg.Text, 150),
				Why:     "First message by this member in the archive.",
			})
		}
	}

	longest := domain.Message{}
	for _, msg := range messages {
		if len(msg.Text) > len(longest.Text) {
			longest = msg
		}
	}
	if longest.ID != "" {
		out = append(out, domain.NotableMessage{
			ID:      longest.ID,
			At:      longest.Timestamp.Format(time.RFC3339),
			Sender:  longest.Sender,
			Kind:    "longest-message",
			Snippet: snippet(longest.Text, 180),
			Why:     "Longest message in the archive.",
		})
	}

	if len(out) > 10 {
		return out[:10]
	}
	return out
}

func fileSHA256(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum[:])
}

func gitCommit() string {
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func providerName(url string) string {
	if url == "" {
		return "heuristic"
	}
	return "ollama"
}

func parameters(input Input) map[string]string {
	out := map[string]string{
		"concurrency": fmt.Sprintf("%d", input.Concurrency),
		"saveEvery":   fmt.Sprintf("%d", input.SaveEvery),
	}
	if !input.Start.IsZero() {
		out["start"] = input.Start.Format("2006-01-02")
	}
	if !input.End.IsZero() {
		out["end"] = input.End.Format("2006-01-02")
	}
	if input.GeneratedAt != "" {
		out["generatedAt"] = input.GeneratedAt
	}
	return out
}

func normalizeWarnings(warnings []domain.Warning) []domain.Warning {
	out := append([]domain.Warning{}, warnings...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Line == out[j].Line {
			if out[i].Code == out[j].Code {
				return out[i].Evidence < out[j].Evidence
			}
			return out[i].Code < out[j].Code
		}
		return out[i].Line < out[j].Line
	})
	return out
}

func cloneStrings(values []string) []string {
	return append([]string{}, values...)
}

func confidence(score float64, evidence []string) domain.Confidence {
	score = roundConfidence(score)
	level := "low"
	switch {
	case score >= 0.8:
		level = "high"
	case score >= 0.6:
		level = "medium"
	}
	clean := make([]string, 0, len(evidence))
	for _, item := range evidence {
		item = strings.TrimSpace(item)
		if item != "" {
			clean = append(clean, item)
		}
	}
	return domain.Confidence{Score: score, Level: level, Evidence: clean}
}

func roundConfidence(score float64) float64 {
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	return float64(int(score*100+0.5)) / 100
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func departureConfidence(gap, span, count int) float64 {
	score := 0.52
	if count > 2 {
		score += 0.12
	}
	if span > 7 {
		score += 0.12
	}
	if gap > 60 {
		score += 0.14
	}
	if gap > 180 {
		score += 0.08
	}
	return score
}

func dedupeInsideJokes(jokes []domain.InsideJoke) []domain.InsideJoke {
	sort.SliceStable(jokes, func(i, j int) bool {
		if jokes[i].Occurrences == jokes[j].Occurrences {
			return len(jokes[i].Phrase) > len(jokes[j].Phrase)
		}
		return jokes[i].Occurrences > jokes[j].Occurrences
	})
	kept := []domain.InsideJoke{}
	for _, joke := range jokes {
		duplicate := false
		for _, existing := range kept {
			if strings.Contains(existing.Phrase, joke.Phrase) || strings.Contains(joke.Phrase, existing.Phrase) {
				if existing.Occurrences == joke.Occurrences || existing.OriginID == joke.OriginID {
					duplicate = true
					break
				}
			}
		}
		if !duplicate {
			kept = append(kept, joke)
		}
	}
	return kept
}

func departureStatus(days int) (string, string) {
	switch {
	case days > 180:
		return "departed", "No recent activity in the observed archive window."
	case days > 60:
		return "quiet", "Activity faded before the archive ended."
	default:
		return "active", "Still active near the end of the archive."
	}
}

func titleFromKeywords(keywords []string) string {
	words := append([]string(nil), keywords...)
	if len(words) > 3 {
		words = words[:3]
	}
	for i := range words {
		words[i] = capitalize(words[i])
	}
	return strings.Join(words, " / ")
}

func capitalize(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func topKeys(counts map[string]int, limit int) []string {
	type pair struct {
		key   string
		count int
	}
	pairs := make([]pair, 0, len(counts))
	for key, count := range counts {
		pairs = append(pairs, pair{key: key, count: count})
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		if pairs[i].count == pairs[j].count {
			return pairs[i].key < pairs[j].key
		}
		return pairs[i].count > pairs[j].count
	})
	if len(pairs) > limit {
		pairs = pairs[:limit]
	}
	out := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		out = append(out, pair.key)
	}
	return out
}

func mentionsName(body, name string) bool {
	pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(strings.ToLower(name)) + `\b`)
	return pattern.MatchString(body)
}

func snippet(text string, limit int) string {
	text = strings.Join(strings.Fields(text), " ")
	if len(text) <= limit {
		return text
	}
	return strings.TrimSpace(text[:limit-1]) + "..."
}

func tooBoring(tokens []string) bool {
	meaningful := 0
	for _, token := range tokens {
		if !stopWords[token] && len(token) > 2 {
			meaningful++
		}
	}
	return meaningful == 0
}

func tokenize(text string) []string {
	normalized := strings.ToLower(text)
	replacer := strings.NewReplacer("'", " ", "\"", " ", "’", " ", "-", " ")
	normalized = replacer.Replace(normalized)
	parts := regexp.MustCompile(`[a-z0-9]+`).FindAllString(normalized, -1)
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) > 1 {
			out = append(out, part)
		}
	}
	return out
}

var stopWords = map[string]bool{
	"about": true, "after": true, "again": true, "also": true, "and": true,
	"are": true, "because": true, "been": true, "but": true, "can": true,
	"could": true, "did": true, "for": true, "from": true, "get": true,
	"had": true, "has": true, "have": true, "how": true, "into": true,
	"just": true, "like": true, "not": true, "now": true, "our": true,
	"out": true, "see": true, "she": true, "that": true, "the": true,
	"then": true, "there": true, "they": true, "this": true, "was": true,
	"we": true, "were": true, "what": true, "when": true, "with": true,
	"you": true, "your": true, "will": true,
}
