package chatparse

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/baditaflorin/group-chat-archaeologist/internal/extract"
)

func BenchmarkRealDataFixtures(b *testing.B) {
	fixtureDir := filepath.Join("..", "..", "test", "fixtures", "realdata")
	inputs, err := filepath.Glob(filepath.Join(fixtureDir, "*.*"))
	if err != nil {
		b.Fatal(err)
	}
	for _, inputPath := range inputs {
		if strings.HasSuffix(inputPath, ".expected.json") {
			continue
		}
		name := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
		extracted, err := extract.Text(context.Background(), inputPath, "")
		if err != nil {
			b.Fatal(err)
		}
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := Parse(extracted.Text); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
