package sqlite

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zk-org/zk/internal/adapter/embedding"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/errors"
	"github.com/zk-org/zk/internal/util/paths"
	strutil "github.com/zk-org/zk/internal/util/strings"
)

// NoteIndex persists note indexing results in the SQLite database.
// It implements the port core.NoteIndex and acts as a facade to the DAOs.
type NoteIndex struct {
	notebookPath string
	db           *DB
	dao          *dao
	logger       util.Logger
	config       core.Config
	embedder     embedding.Provider
	embedderErr  error
	embedderInit bool
}

type dao struct {
	notes       *NoteDAO
	links       *LinkDAO
	collections *CollectionDAO
	metadata    *MetadataDAO
}

func NewNoteIndex(notebookPath string, db *DB, logger util.Logger, config core.Config) *NoteIndex {
	return &NoteIndex{
		notebookPath: notebookPath,
		db:           db,
		logger:       logger,
		config:       config,
	}
}

// Find implements core.NoteIndex.
func (ni *NoteIndex) Find(opts core.NoteFindOpts) (notes []core.ContextualNote, err error) {
	err = ni.commit(func(dao *dao) error {
		if opts.MatchStrategy == core.MatchStrategyNL {
			embedder, err := ni.getEmbedder()
			if err != nil {
				return err
			}
			notes, err = dao.notes.FindNaturalLanguage(opts, embedder, ni.config.Embedding)
			return err
		}
		notes, err = dao.notes.Find(opts)
		return err
	})
	return
}

// FindMinimal implements core.NoteIndex.
func (ni *NoteIndex) FindMinimal(opts core.NoteFindOpts) (notes []core.MinimalNote, err error) {
	err = ni.commit(func(dao *dao) error {
		if opts.MatchStrategy == core.MatchStrategyNL {
			embedder, err := ni.getEmbedder()
			if err != nil {
				return err
			}
			notes, err = dao.notes.FindMinimalNaturalLanguage(opts, embedder, ni.config.Embedding)
			return err
		}
		notes, err = dao.notes.FindMinimal(opts)
		return err
	})
	return
}

// FindLinkMatch implements core.NoteIndex.
func (ni *NoteIndex) FindLinkMatch(baseDir string, href string, linkType core.LinkType) (id core.NoteID, err error) {
	err = ni.commit(func(dao *dao) error {
		id, err = ni.findLinkMatch(dao, baseDir, href, linkType)
		return err
	})
	return
}

func (ni *NoteIndex) findLinkMatch(dao *dao, baseDir string, href string, linkType core.LinkType) (core.NoteID, error) {
	if strutil.IsURL(href) {
		return 0, nil
	}

	id, _ := ni.findPathMatch(dao, baseDir, href)
	if id.IsValid() {
		return id, nil
	}

	allowPartialMatch := (linkType == core.LinkTypeWikiLink)
	return dao.notes.FindIDByHref(href, allowPartialMatch)
}

func (ni *NoteIndex) findPathMatch(dao *dao, baseDir string, href string) (core.NoteID, error) {
	href, err := ni.relNotebookPath(baseDir, href)
	if err != nil {
		return 0, err
	}
	return dao.notes.FindIDByHref(href, false)
}

// FindLinksBetweenNotes implements core.NoteIndex.
func (ni *NoteIndex) FindLinksBetweenNotes(ids []core.NoteID) (links []core.ResolvedLink, err error) {
	err = ni.commit(func(dao *dao) error {
		links, err = dao.links.FindBetweenNotes(ids)
		return err
	})
	return
}

// FindCollections implements core.NoteIndex.
func (ni *NoteIndex) FindCollections(kind core.CollectionKind, sorters []core.CollectionSorter) (collections []core.Collection, err error) {
	err = ni.commit(func(dao *dao) error {
		collections, err = dao.collections.FindAll(kind, sorters)
		return err
	})
	return
}

// IndexedPaths implements core.NoteIndex.
func (ni *NoteIndex) IndexedPaths() (metadata <-chan paths.Metadata, err error) {
	err = ni.commit(func(dao *dao) error {
		metadata, err = dao.notes.Indexed()
		return err
	})
	err = errors.Wrap(err, "failed to get indexed notes")
	return
}

// Add implements core.NoteIndex.
func (ni *NoteIndex) Add(note core.Note) (id core.NoteID, err error) {
	err = ni.commit(func(dao *dao) error {
		id, err = dao.notes.Add(note)
		if err != nil {
			return err
		}
		note.ID = id

		err = ni.addLinks(dao, id, note.Links)
		if err != nil {
			return err
		}

		err = ni.fixExistingLinks(dao, note.ID, note.Path)
		if err != nil {
			return err
		}

		err = ni.associateTags(dao.collections, id, note.Tags)
		if err != nil {
			return err
		}

		return ni.indexNoteEmbeddings(dao.notes, id, note)
	})

	err = errors.Wrapf(err, "%v: failed to index the note", note.Path)
	return
}

// fixExistingLinks will go over all indexed links and update their target to
// the given id if they match the given path better than their current
// targetPath.
func (ni *NoteIndex) fixExistingLinks(dao *dao, id core.NoteID, path string) error {
	links, err := dao.links.FindInternal()
	if err != nil {
		return err
	}

	for _, link := range links {
		// To find the best match possible, shortest paths take precedence.
		// See https://github.com/zk-org/zk/issues/23
		if link.TargetPath != "" && len(link.TargetPath) < len(path) {
			continue
		}

		// FIXME: err is never cheked
		if matches, err := ni.linkMatchesPath(link, path); matches && err == nil {
			err = dao.links.SetTargetID(link.ID, id)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// linkMatchesPath returns whether the given link can be used to reach the
// given note path.
func (ni *NoteIndex) linkMatchesPath(link core.ResolvedLink, path string) (bool, error) {
	// Remove any anchor at the end of the HREF, since it's most likely
	// matching a sub-section in the note.
	href := strings.SplitN(link.Href, "#", 2)[0]

	matchString := func(pattern string, s string) bool {
		reg := regexp.MustCompile(pattern)
		return reg.MatchString(s)
	}

	matches := func(href string, allowPartialHref bool) bool {
		if href == "" {
			return false
		}
		href = regexp.QuoteMeta(href)

		if allowPartialHref {
			if matchString("^(.*/)?[^/]*"+href+"[^/]*$", path) {
				return true
			}
			if matchString(".*"+href+".*", path) {
				return true
			}
		}

		return matchString("^(?:"+href+"[^/]*|"+href+"/.+)$", path)
	}

	baseDir := filepath.Join(ni.notebookPath, filepath.Dir(link.SourcePath))
	if relHref, err := ni.relNotebookPath(baseDir, href); err != nil {
		if matches(relHref, false) {
			return true, nil
		}
	}

	allowPartialMatch := (link.Type == core.LinkTypeWikiLink)
	return matches(href, allowPartialMatch), nil
}

// relNotebookHref makes the given href (which is relative to baseDir) relative
// to the notebook root instead.
func (ni *NoteIndex) relNotebookPath(baseDir string, href string) (string, error) {
	path := filepath.Clean(filepath.Join(baseDir, href))
	path, err := filepath.Rel(ni.notebookPath, path)

	return path,
		errors.Wrapf(err, "failed to make href relative to the notebook: %s", href)
}

// Update implements core.NoteIndex.
func (ni *NoteIndex) Update(note core.Note) error {
	err := ni.commit(func(dao *dao) error {
		id, err := dao.notes.Update(note)
		if err != nil {
			return err
		}

		// Reset links
		err = dao.links.RemoveAll(id)
		if err != nil {
			return err
		}
		err = ni.addLinks(dao, id, note.Links)
		if err != nil {
			return err
		}

		// Reset tags
		err = dao.collections.RemoveAssociations(id)
		if err != nil {
			return err
		}
		err = ni.associateTags(dao.collections, id, note.Tags)
		if err != nil {
			return err
		}

		return ni.indexNoteEmbeddings(dao.notes, id, note)
	})

	return errors.Wrapf(err, "%v: failed to update note index", note.Path)
}

func (ni *NoteIndex) associateTags(collections *CollectionDAO, noteID core.NoteID, tags []string) error {
	for _, tag := range tags {
		tagID, err := collections.FindOrCreate(core.CollectionKindTag, tag)
		if err != nil {
			return err
		}
		_, err = collections.Associate(noteID, tagID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ni *NoteIndex) addLinks(dao *dao, id core.NoteID, links []core.Link) error {
	resolvedLinks, err := ni.resolveLinkNoteIDs(dao, id, links)
	if err != nil {
		return err
	}
	return dao.links.Add(resolvedLinks)
}

func (ni *NoteIndex) resolveLinkNoteIDs(dao *dao, sourceID core.NoteID, links []core.Link) ([]core.ResolvedLink, error) {
	resolvedLinks := []core.ResolvedLink{}

	for _, link := range links {
		targetID, err := ni.findLinkMatch(dao, "" /* base dir */, link.Href, link.Type)
		if err != nil {
			return resolvedLinks, err
		}

		resolvedLinks = append(resolvedLinks, core.ResolvedLink{
			Link:     link,
			SourceID: sourceID,
			TargetID: targetID,
		})
	}

	return resolvedLinks, nil
}

// Remove implements core.NoteIndex
func (ni *NoteIndex) Remove(path string) error {
	err := ni.commit(func(dao *dao) error {
		id, err := dao.notes.FindIDByPath(path)
		if err != nil {
			return err
		}
		if id.IsValid() && ni.config.Embedding.Enabled {
			if err := dao.notes.RemoveEmbeddings(id); err != nil {
				return err
			}
		}
		return dao.notes.Remove(path)
	})
	return errors.Wrapf(err, "%v: failed to remove note from index", path)
}

// Commit implements core.NoteIndex.
func (ni *NoteIndex) Commit(transaction func(idx core.NoteIndex) error) error {
	return ni.commit(func(dao *dao) error {
		return transaction(&NoteIndex{
			notebookPath: ni.notebookPath,
			db:           ni.db,
			dao:          dao,
			logger:       ni.logger,
			config:       ni.config,
			embedder:     ni.embedder,
		})
	})
}

// NeedsReindexing implements core.NoteIndex.
func (ni *NoteIndex) NeedsReindexing() (needsReindexing bool, err error) {
	err = ni.commit(func(dao *dao) error {
		res, err := dao.metadata.Get(reindexingRequiredKey)
		needsReindexing = (res == "true")
		if err != nil {
			return err
		}

		if ni.config.Embedding.Enabled {
			expected := ni.embeddingConfigSignature()
			actual, err := dao.metadata.Get(embeddingConfigSignatureKey)
			if err != nil {
				return err
			}
			if actual != expected {
				needsReindexing = true
			}
		}
		return nil
	})
	return
}

// SetNeedsReindexing implements core.NoteIndex.
func (ni *NoteIndex) SetNeedsReindexing(needsReindexing bool) error {
	return ni.commit(func(dao *dao) error {
		value := "false"
		if needsReindexing {
			value = "true"
		}

		if err := dao.metadata.Set(reindexingRequiredKey, value); err != nil {
			return err
		}

		if !needsReindexing && ni.config.Embedding.Enabled {
			return dao.metadata.Set(embeddingConfigSignatureKey, ni.embeddingConfigSignature())
		}
		return nil
	})
}

func (ni *NoteIndex) commit(transaction func(dao *dao) error) error {
	if ni.dao != nil {
		return transaction(ni.dao)
	} else {
		return ni.db.WithTransaction(func(tx Transaction) error {
			dao := dao{
				notes:       NewNoteDAO(tx, ni.logger),
				links:       NewLinkDAO(tx, ni.logger),
				collections: NewCollectionDAO(tx, ni.logger),
				metadata:    NewMetadataDAO(tx),
			}
			return transaction(&dao)
		})
	}
}

func (ni *NoteIndex) embeddingConfigSignature() string {
	raw := fmt.Sprintf(
		"%t|%s|%s|%s|%d|%d|%d|%d",
		ni.config.Embedding.Enabled,
		ni.config.Embedding.Provider,
		ni.config.Embedding.Model,
		ni.config.Embedding.Endpoint,
		ni.config.Embedding.Dimensions,
		ni.config.Embedding.ChunkSize,
		ni.config.Embedding.ChunkOverlap,
		ni.config.Embedding.MaxChunksPerNote,
	)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (ni *NoteIndex) getEmbedder() (embedding.Provider, error) {
	if !ni.config.Embedding.Enabled {
		return nil, fmt.Errorf("embedding search is disabled; set [embedding].enabled = true in config.toml")
	}
	if !vecBuildEnabled() {
		return nil, fmt.Errorf("semantic search requires zk built with -tags vec")
	}
	if ni.embedderInit {
		return ni.embedder, ni.embedderErr
	}

	ni.embedder, ni.embedderErr = embedding.NewProvider(ni.config.Embedding)
	ni.embedderInit = true
	return ni.embedder, ni.embedderErr
}

func (ni *NoteIndex) indexNoteEmbeddings(notes *NoteDAO, id core.NoteID, note core.Note) error {
	if !ni.config.Embedding.Enabled {
		return nil
	}

	embedder, err := ni.getEmbedder()
	if err != nil {
		return err
	}

	return notes.ReindexEmbeddings(
		context.Background(),
		id,
		note,
		embedder,
		ni.config.Embedding,
	)
}
