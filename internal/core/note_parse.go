package core

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/relvacode/iso8601"
	"github.com/zk-org/zk/internal/util/errors"
	"github.com/zk-org/zk/internal/util/opt"
	strutil "github.com/zk-org/zk/internal/util/strings"
	"gopkg.in/djherbis/times.v1"
)

// NoteParser parses a note on the file system into a Note model.
type NoteParser interface {
	ParseNoteAt(absPath string) (*Note, error)
}

// NoteContentParser parses a note's raw content into its components.
type NoteContentParser interface {
	ParseNoteContent(content string) (*NoteContent, error)
}

// NoteContent holds the data parsed from the note content.
type NoteContent struct {
	// Title is the heading of the note.
	Title opt.String
	// Lead is the opening paragraph or section of the note.
	Lead opt.String
	// Body is the content of the note, including the Lead but without the Title.
	Body opt.String
	// Tags is the list of tags found in the note content.
	Tags []string
	// Links is the list of outbound links found in the note.
	Links []Link
	// Additional metadata. For example, extracted from a YAML frontmatter.
	Metadata map[string]interface{}
}

// ParseNoteAt implements NoteParser.
func (n *Notebook) ParseNoteAt(absPath string) (*Note, error) {
	wrap := errors.Wrapper(absPath)

	content, err := n.fs.Read(absPath)
	if err != nil {
		return nil, wrap(err)
	}

	return n.ParseNoteWithContent(absPath, content)
}

func (n *Notebook) ParseNoteWithContent(absPath string, content []byte) (*Note, error) {
	wrap := errors.Wrapper(absPath)

	relPath, err := n.RelPath(absPath)
	if err != nil {
		return nil, wrap(err)
	}

	contentStr := string(content)
	contentParts, err := n.Parser.ParseNoteContent(contentStr)
	if err != nil {
		return nil, wrap(err)
	}

	note := Note{
		Path:       relPath,
		Title:      contentParts.Title.String(),
		Lead:       contentParts.Lead.String(),
		Body:       contentParts.Body.String(),
		RawContent: contentStr,
		WordCount:  len(strings.Fields(contentStr)),
		Links:      make([]Link, 0),
		Tags:       contentParts.Tags,
		Metadata:   contentParts.Metadata,
		Checksum:   fmt.Sprintf("%x", sha256.Sum256(content)),
	}

	for _, link := range contentParts.Links {
		if !strutil.IsURL(link.Href) && link.Type == LinkTypeMarkdown {
			// Make the href relative to the notebook root.
			href := filepath.Join(filepath.Dir(absPath), link.Href)
			link.Href, err = n.RelPath(href)
			if err != nil {
				n.logger.Err(err)
				continue
			}
		}
		note.Links = append(note.Links, link)
	}

	times, err := times.Stat(absPath)
	if err == nil {
		note.Modified = times.ModTime().UTC()
		note.Created = creationDateFrom(note.Metadata, times)
	}

	return &note, nil
}

func creationDateFrom(metadata map[string]interface{}, times times.Timespec) time.Time {
	// Read the creation date from the YAML frontmatter `date` key.
	if dateVal, ok := metadata["date"]; ok {
		if dateStr, ok := dateVal.(string); ok {
			if time, err := iso8601.ParseString(dateStr); err == nil {
				return time
			}
			// Omitting the `T` is common
			if time, err := time.Parse("2006-01-02 15:04:05", dateStr); err == nil {
				return time
			}
			if time, err := time.Parse("2006-01-02 15:04", dateStr); err == nil {
				return time
			}
		}
	}

	if times.HasBirthTime() {
		return times.BirthTime().UTC()
	}

	return time.Now().UTC()
}
