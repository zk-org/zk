// gentestdata generates a test notebook with many notes for benchmarking.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

var words = []string{
	"the", "be", "to", "of", "and", "a", "in", "that", "have", "I",
	"it", "for", "not", "on", "with", "he", "as", "you", "do", "at",
	"this", "but", "his", "by", "from", "they", "we", "say", "her", "she",
	"or", "an", "will", "my", "one", "all", "would", "there", "their", "what",
	"so", "up", "out", "if", "about", "who", "get", "which", "go", "me",
	"when", "make", "can", "like", "time", "no", "just", "him", "know", "take",
	"people", "into", "year", "your", "good", "some", "could", "them", "see", "other",
	"than", "then", "now", "look", "only", "come", "its", "over", "think", "also",
	"back", "after", "use", "two", "how", "our", "work", "first", "well", "way",
	"even", "new", "want", "because", "any", "these", "give", "day", "most", "us",
	"note", "idea", "thought", "concept", "link", "reference", "knowledge", "learning",
	"writing", "reading", "research", "study", "analysis", "synthesis", "connection",
}

func main() {
	var (
		outputDir = flag.String("output", "testdata", "output directory for test notebook")
		noteCount = flag.Int("notes", 3000, "number of notes to generate")
		seed      = flag.Int64("seed", 42, "random seed for reproducibility")
	)
	flag.Parse()

	rng := rand.New(rand.NewSource(*seed))

	if err := generateNotebook(*outputDir, *noteCount, rng); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %d notes in %s using seed %d\n", *noteCount, *outputDir, seed)
}

func generateNotebook(outputDir string, noteCount int, rng *rand.Rand) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	zkDir := filepath.Join(outputDir, ".zk")
	if err := os.MkdirAll(zkDir, 0755); err != nil {
		return fmt.Errorf("creating .zk directory: %w", err)
	}

	for i := range noteCount {
		filename := fmt.Sprintf("note-%04d.md", i)
		path := filepath.Join(outputDir, filename)

		content := generateNote(i, noteCount, rng)

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", filename, err)
		}
	}

	return nil
}

func generateNote(noteIndex, totalNotes int, rng *rand.Rand) string {
	var content strings.Builder
	fmt.Fprintf(&content, "# Test note %d\n\n", noteIndex)

	paragraphCount := 2 + rng.Intn(3)
	for range paragraphCount {
		content.WriteString(generateParagraph(rng) + "\n\n")
	}

	linkCount := 3 + rng.Intn(4)
	content.WriteString("## Related\n\n")

	usedTargets := make(map[int]bool)
	for range linkCount {
		var target int
		for {
			target = rng.Intn(totalNotes)
			if target != noteIndex && !usedTargets[target] {
				usedTargets[target] = true
				break
			}
		}

		linkTitle := generateLinkTitle(rng)
		fmt.Fprintf(&content, "- [%s](note-%04d.md)\n", linkTitle, target)
	}

	return content.String()
}

func generateParagraph(rng *rand.Rand) string {
	sentenceCount := 3 + rng.Intn(4)
	var paragraph strings.Builder

	for s := range sentenceCount {
		if s > 0 {
			paragraph.WriteString(" ")
		}
		paragraph.WriteString(generateSentence(rng))
	}

	return paragraph.String()
}

func generateSentence(rng *rand.Rand) string {
	wordCount := 5 + rng.Intn(8)
	var sentence strings.Builder

	for w := range wordCount {
		word := words[rng.Intn(len(words))]
		if w == 0 {
			word = string(word[0]-32) + word[1:]
		}
		if w > 0 {
			sentence.WriteString(" ")
		}
		sentence.WriteString(word)
	}

	return sentence.String() + "."
}

func generateLinkTitle(rng *rand.Rand) string {
	wordCount := 2 + rng.Intn(3)
	var title strings.Builder

	for w := range wordCount {
		if w > 0 {
			title.WriteString(" ")
		}
		title.WriteString(words[rng.Intn(len(words))])
	}

	return title.String()
}
