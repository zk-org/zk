package file

// DiffChange represents a note change made in a slip box directory.
type DiffChange struct {
	Path Path
	Kind DiffKind
}

// DiffKind represents a type of note change made in a slip box directory.
type DiffKind int

const (
	DiffAdded DiffKind = iota + 1
	DiffModified
	DiffRemoved
)

// Diff compares two sources of FileMetadata and report the file changes, using
// the file modification date.
//
// Warning: The FileMetadata have to be sorted by their Path for the diffing to
// work properly.
func Diff(source, target <-chan Metadata, callback func(DiffChange) error) error {
	var err error
	var sourceFile, targetFile Metadata
	var sourceOpened, targetOpened bool = true, true
	pair := diffPair{}

	for err == nil && (sourceOpened || targetOpened) {
		if pair.source == nil {
			sourceFile, sourceOpened = <-source
			if sourceOpened {
				pair.source = &sourceFile
			}
		}
		if pair.target == nil {
			targetFile, targetOpened = <-target
			if targetOpened {
				pair.target = &targetFile
			}
		}
		change := pair.diff()
		if change != nil {
			err = callback(*change)
		}
	}

	return err
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
func (p *diffPair) diff() *DiffChange {
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
		if p.source.Modified != p.target.Modified {
			change = &DiffChange{p.source.Path, DiffModified}
		}
		p.source = nil
		p.target = nil

	default: // Different files, one has been added or removed.
		if p.source.Path.Less(p.target.Path) {
			change = &DiffChange{p.source.Path, DiffAdded}
			p.source = nil
		} else {
			change = &DiffChange{p.target.Path, DiffRemoved}
			p.target = nil
		}
	}

	return change
}
