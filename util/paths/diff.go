package paths

import "fmt"

// DiffChange represents a file change made in a directory.
type DiffChange struct {
	Path string
	Kind DiffKind
}

// String implements Stringer.
func (c DiffChange) String() string {
	return fmt.Sprintf("%v %v", c.Kind, c.Path)
}

// DiffKind represents a type of file change made in a directory.
type DiffKind int

const (
	DiffAdded DiffKind = iota + 1
	DiffModified
	DiffRemoved
)

// String implements Stringer.
func (k DiffKind) String() string {
	switch k {
	case DiffAdded:
		return "+"
	case DiffModified:
		return "~"
	case DiffRemoved:
		return "-"
	default:
		panic(fmt.Sprintf("%d: unknown DiffKind", int(k)))
	}
}

// Diff compares two sources of Metadata and report the file changes, using the
// file modification date.
//
// Returns the number of files in the source.
//
// Warning: The Metadata have to be sorted by their Path for the diffing to
// work properly.
func Diff(source, target <-chan Metadata, forceModified bool, callback func(DiffChange) error) (int, error) {
	var err error
	var sourceFile, targetFile Metadata
	var sourceOpened, targetOpened bool = true, true
	sourceCount := 0
	pair := diffPair{}

	for err == nil && (sourceOpened || targetOpened) {
		if pair.source == nil {
			sourceFile, sourceOpened = <-source
			if sourceOpened {
				sourceCount += 1
				pair.source = &sourceFile
			}
		}
		if pair.target == nil {
			targetFile, targetOpened = <-target
			if targetOpened {
				pair.target = &targetFile
			}
		}
		change := pair.diff(forceModified)
		if change != nil {
			err = callback(*change)
		}
	}

	return sourceCount, err
}

// diffPair holds the current two files to be diffed.
type diffPair struct {
	source *Metadata
	target *Metadata
}

// diff compares the source and target files in the current pair.
//
// If the source and target file are at the same path, we check for any change.
// If the files are different, that means that either the source file was
// added, or the target file was removed.
func (p *diffPair) diff(forceModified bool) *DiffChange {
	var change *DiffChange

	switch {
	case p.source == nil && p.target == nil: // Both channels are closed
		break

	case p.source == nil && p.target != nil: // Source channel is closed
		change = &DiffChange{p.target.Path, DiffRemoved}
		p.target = nil

	case p.source != nil && p.target == nil: // Target channel is closed
		change = &DiffChange{p.source.Path, DiffAdded}
		p.source = nil

	case p.source.Path == p.target.Path: // Same files, compare their modification date.
		if forceModified || p.source.Modified != p.target.Modified {
			change = &DiffChange{p.source.Path, DiffModified}
		}
		p.source = nil
		p.target = nil

	default: // Different files, one has been added or removed.
		if p.source.Path < p.target.Path {
			change = &DiffChange{p.source.Path, DiffAdded}
			p.source = nil
		} else {
			change = &DiffChange{p.target.Path, DiffRemoved}
			p.target = nil
		}
	}

	return change
}
