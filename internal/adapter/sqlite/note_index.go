package sqlite

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/paths"
	strutil "github.com/zk-org/zk/internal/util/strings"
	"golang.org/x/sync/errgroup"
)

// NoteIndex persists note indexing results in the SQLite database.
// It implements the port core.NoteIndex and acts as a facade to the DAOs.
type NoteIndex struct {
	notebookPath string
	db           *DB
	dao          *dao
	logger       util.Logger
	extension    string
}

type dao struct {
	notes       *NoteDAO
	links       *LinkDAO
	collections *CollectionDAO
	metadata    *MetadataDAO
}

func NewNoteIndex(notebookPath string, db *DB, logger util.Logger, extension string) *NoteIndex {
	return &NoteIndex{
		notebookPath: notebookPath,
		db:           db,
		logger:       logger,
		extension:    extension,
	}
}

// Find implements core.NoteIndex.
func (ni *NoteIndex) Find(opts core.NoteFindOpts) (notes []core.ContextualNote, err error) {
	err = ni.commit(func(dao *dao) error {
		notes, err = dao.notes.Find(opts)
		return err
	})
	return
}

// FindMinimal implements core.NoteIndex.
func (ni *NoteIndex) FindMinimal(opts core.NoteFindOpts) (notes []core.MinimalNote, err error) {
	err = ni.commit(func(dao *dao) error {
		notes, err = dao.notes.FindMinimal(opts)
		return err
	})
	return
}

func (ni *NoteIndex) findLinkMatch(dao *dao, baseDir string, href string, linkType core.LinkType) (core.NoteID, error) {
	if strutil.IsURL(href) {
		return 0, nil
	}

	// Try exact path match first
	pathHref, err := ni.relNotebookPath(baseDir, href)
	if err == nil {
		id, _ := dao.notes.FindIDByHref(pathHref, false)
		if id.IsValid() {
			return id, nil
		}
	}

	// Fall back to partial matching for wiki links
	allowPartialMatch := (linkType == core.LinkTypeWikiLink)
	return dao.notes.FindIDByHref(href, allowPartialMatch)
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
	if err != nil {
		err = fmt.Errorf("failed to get indexed notes: %w", err)
	}
	return
}

// Add implements core.NoteIndex.
func (ni *NoteIndex) Add(note core.Note, fixLinks bool) (id core.NoteID, err error) {
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

		if fixLinks {
			err = ni.fixExistingLinks(dao, note.ID, note.Path)
			if err != nil {
				return err
			}
		}

		return ni.associateTags(dao.collections, id, note.Tags)
	})

	if err != nil {
		err = fmt.Errorf("%v: failed to index the note: %w", note.Path, err)
	}
	return
}

func (ni *NoteIndex) fixExistingLinks(dao *dao, id core.NoteID, path string) error {
	return ni.batchFixExistingLinks(dao, []core.NoteID{id}, []string{path})
}

// BatchUpdateLinks will go over all indexed links and update their target to
// one of the given ids if its path better matches their current targetPath.
func (ni *NoteIndex) BatchUpdateLinks(ids []core.NoteID, paths []string) error {
	return ni.commit(func(dao *dao) error {
		return ni.batchFixExistingLinks(dao, ids, paths)
	})
}

func (ni *NoteIndex) batchFixExistingLinks(dao *dao, ids []core.NoteID, paths []string) error {
	links, err := dao.links.FindInternal()
	if err != nil || len(links) == 0 {
		return err
	}

	fixLink := func(link core.ResolvedLink) error {
		bestTargetPath := link.TargetPath
		bestMatch := -1
		for i, path := range paths {
			// To find the best match possible, shortest paths take precedence.
			// See https://github.com/zk-org/zk/issues/23
			if bestTargetPath != "" && len(bestTargetPath) < len(path) {
				continue
			}

			matches, err := ni.linkMatchesPath(link, path)
			if err != nil {
				return err
			}
			if matches {
				bestTargetPath = path
				bestMatch = i
			}
		}

		if bestMatch != -1 {
			err = dao.links.SetTargetID(link.ID, ids[bestMatch])
			if err != nil {
				return err
			}
		}

		return nil
	}

	// Update the links in parallel: we must parse through the links of all notes,
	// which is a considerable amount of work. We do so without using all CPU cores at once.
	maxWorkers := min(max(runtime.GOMAXPROCS(0)-2, 1), len(links))
	linksPerWorker := len(links) / maxWorkers
	group := new(errgroup.Group)
	for wid := range maxWorkers {
		start := linksPerWorker * wid
		end := linksPerWorker * (wid + 1)
		if wid == maxWorkers-1 {
			// The last worker does a bit more work
			end += len(links) % maxWorkers
		}
		group.Go(func() error {
			for _, link := range links[start:end] {
				err := fixLink(link)
				if err != nil {
					return err
				}
			}
			return nil
		})
	}
	return group.Wait()
}

// linkMatchesPath returns whether the given link can be used to reach the
// given note path.
func (ni *NoteIndex) linkMatchesPath(link core.ResolvedLink, path string) (bool, error) {
	// Remove any anchor at the end of the HREF, since it's most likely
	// matching a sub-section in the note.
	href := link.Href
	if hashPos := strings.IndexByte(link.Href, '#'); hashPos != -1 {
		href = link.Href[:hashPos]
	}

	matches := func(href string, allowPartialHref bool) bool {
		if href == "" {
			return false
		}
		pos := strings.Index(path, href)
		if allowPartialHref && pos != -1 {
			// Match if 'href' is anywhere in 'path'
			return true
		} else if pos != 0 {
			// Otherwise 'path' must start with 'href'
			return false
		}

		slashPos := strings.IndexByte(path[len(href):], '/')
		if slashPos != -1 {
			// 'href/abc', 'href/a/b/c', or 'href/' but not 'hrefAnd/something/after'
			return slashPos == 0
		}

		// 'href' or 'hrefSomeSuffix'
		return true
	}

	allowPartialMatch := link.Type == core.LinkTypeWikiLink
	if matches(href, allowPartialMatch) {
		return true, nil
	}

	baseDir := filepath.Join(ni.notebookPath, filepath.Dir(link.SourcePath))
	if relHref, err := ni.relNotebookPath(baseDir, href); err == nil {
		if matches(relHref, false) {
			return true, nil
		}
	}

	return false, nil
}

// relNotebookHref makes the given href (which is relative to baseDir) relative
// to the notebook root instead.
func (ni *NoteIndex) relNotebookPath(baseDir string, href string) (string, error) {
	path := filepath.Clean(filepath.Join(baseDir, href))
	path, err := filepath.Rel(ni.notebookPath, path)

	if err != nil {
		return "", fmt.Errorf("failed to make href relative to the notebook: %s: %w", href, err)
	}
	return path, nil
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
		return ni.associateTags(dao.collections, id, note.Tags)
	})

	if err != nil {
		return fmt.Errorf("%v: failed to update note index: %w", note.Path, err)
	}
	return nil
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
		return dao.notes.Remove(path)
	})
	if err != nil {
		return fmt.Errorf("%v: failed to remove note from index: %w", path, err)
	}
	return nil
}

// Commit implements core.NoteIndex.
func (ni *NoteIndex) Commit(transaction func(idx core.NoteIndex) error) error {
	return ni.commit(func(dao *dao) error {
		return transaction(&NoteIndex{
			db:     ni.db,
			dao:    dao,
			logger: ni.logger,
		})
	})
}

// NeedsReindexing implements core.NoteIndex.
func (ni *NoteIndex) NeedsReindexing() (needsReindexing bool, err error) {
	err = ni.commit(func(dao *dao) error {
		res, err := dao.metadata.Get(reindexingRequiredKey)
		needsReindexing = (res == "true")
		return err
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

		return dao.metadata.Set(reindexingRequiredKey, value)
	})
}

func (ni *NoteIndex) commit(transaction func(dao *dao) error) error {
	if ni.dao != nil {
		return transaction(ni.dao)
	} else {
		return ni.db.WithTransaction(func(tx Transaction) error {
			dao := dao{
				notes:       NewNoteDAO(tx, ni.logger, ni.extension),
				links:       NewLinkDAO(tx, ni.logger),
				collections: NewCollectionDAO(tx, ni.logger),
				metadata:    NewMetadataDAO(tx),
			}
			return transaction(&dao)
		})
	}
}
