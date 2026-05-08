package storage

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/baditaflorin/group-chat-archaeologist/internal/domain"
)

type Store struct {
	duckdbPath string
}

func Open() (*Store, error) {
	path, err := exec.LookPath("duckdb")
	if err != nil {
		return &Store{}, nil
	}
	return &Store{duckdbPath: path}, nil
}

func (s *Store) Close() error {
	return nil
}

func (s *Store) Summarize(ctx context.Context, messages []domain.Message) (domain.StorageSummary, error) {
	if s.duckdbPath == "" {
		return fallbackSummary(messages), nil
	}

	summary, err := s.duckDBSummary(ctx, messages)
	if err != nil {
		fallback := fallbackSummary(messages)
		fallback.Engine = "go-fallback-after-duckdb-error"
		return fallback, nil
	}
	return summary, nil
}

func (s *Store) duckDBSummary(ctx context.Context, messages []domain.Message) (domain.StorageSummary, error) {
	tmpDir, err := os.MkdirTemp("", "group-chat-archaeologist-duckdb-*")
	if err != nil {
		return domain.StorageSummary{}, fmt.Errorf("create duckdb temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	csvPath := filepath.Join(tmpDir, "messages.csv")
	if err := writeMessagesCSV(csvPath, messages); err != nil {
		return domain.StorageSummary{}, err
	}

	query := fmt.Sprintf(`
		SELECT
			sender AS name,
			COUNT(*)::INTEGER AS message_count,
			MIN(CAST(ts AS TIMESTAMP)) AS first_message_at,
			MAX(CAST(ts AS TIMESTAMP)) AS last_message_at
		FROM read_csv_auto('%s', header=true)
		GROUP BY sender
		ORDER BY message_count DESC, sender ASC
	`, strings.ReplaceAll(csvPath, "'", "''"))

	cmd := exec.CommandContext(ctx, s.duckdbPath, "-json", "-c", query)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return domain.StorageSummary{}, fmt.Errorf("run duckdb: %w: %s", err, strings.TrimSpace(string(output)))
	}

	var rows []struct {
		Name           string `json:"name"`
		MessageCount   int    `json:"message_count"`
		FirstMessageAt string `json:"first_message_at"`
		LastMessageAt  string `json:"last_message_at"`
	}
	if err := json.Unmarshal(output, &rows); err != nil {
		return domain.StorageSummary{}, fmt.Errorf("parse duckdb JSON: %w", err)
	}

	stats := make([]domain.MemberStat, 0, len(rows))
	for _, row := range rows {
		first, err := parseDuckTime(row.FirstMessageAt)
		if err != nil {
			return domain.StorageSummary{}, err
		}
		last, err := parseDuckTime(row.LastMessageAt)
		if err != nil {
			return domain.StorageSummary{}, err
		}
		stats = append(stats, domain.MemberStat{
			Name:           row.Name,
			MessageCount:   row.MessageCount,
			FirstMessageAt: first,
			LastMessageAt:  last,
		})
	}

	return domain.StorageSummary{MemberStats: stats, Engine: "duckdb-cli"}, nil
}

func writeMessagesCSV(path string, messages []domain.Message) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create messages csv: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"id", "ts", "sender", "body"}); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}
	for _, msg := range messages {
		if err := writer.Write([]string{
			msg.ID,
			msg.Timestamp.Format(time.RFC3339),
			msg.Sender,
			msg.Text,
		}); err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("flush csv: %w", err)
	}
	return nil
}

func fallbackSummary(messages []domain.Message) domain.StorageSummary {
	statsByMember := map[string]domain.MemberStat{}
	for _, msg := range messages {
		stat, ok := statsByMember[msg.Sender]
		if !ok {
			stat = domain.MemberStat{Name: msg.Sender, FirstMessageAt: msg.Timestamp, LastMessageAt: msg.Timestamp}
		}
		stat.MessageCount++
		if msg.Timestamp.Before(stat.FirstMessageAt) {
			stat.FirstMessageAt = msg.Timestamp
		}
		if msg.Timestamp.After(stat.LastMessageAt) {
			stat.LastMessageAt = msg.Timestamp
		}
		statsByMember[msg.Sender] = stat
	}

	stats := make([]domain.MemberStat, 0, len(statsByMember))
	for _, stat := range statsByMember {
		stats = append(stats, stat)
	}
	sort.SliceStable(stats, func(i, j int) bool {
		if stats[i].MessageCount == stats[j].MessageCount {
			return stats[i].Name < stats[j].Name
		}
		return stats[i].MessageCount > stats[j].MessageCount
	})
	return domain.StorageSummary{MemberStats: stats, Engine: "go-fallback"}
}

func parseDuckTime(value string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.999999",
	}
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, value, time.UTC)
		if err == nil {
			return parsed.UTC(), nil
		}
	}
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		return time.Unix(seconds, 0).UTC(), nil
	}
	return time.Time{}, fmt.Errorf("parse duckdb timestamp %q", value)
}
