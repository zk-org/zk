package note

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

// Diff compares two sources of FileMetadata and report the note changes, using
// the file checksum and modification date.
//
// Warning: The FileMetadata have to be sorted by their Path for the diffing to
// work properly.
func Diff(source, target <-chan FileMetadata) <-chan DiffChange {
	c := make(chan DiffChange)
	go func() {
		defer close(c)

		pair := diffPair{}
		var sourceFile, targetFile FileMetadata
		var sourceOpened, targetOpened bool = true, true

		for sourceOpened || targetOpened {
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
				c <- *change
			}
		}
	}()
	return c
}

// diffPair holds the current two files to be diffed.
type diffPair struct {
	source *FileMetadata
	target *FileMetadata
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
		if isAscendingOrder(p.source.Path, p.target.Path) {
			change = &DiffChange{p.source.Path, DiffAdded}
			p.source = nil
		} else {
			change = &DiffChange{p.target.Path, DiffRemoved}
			p.target = nil
		}
	}

	return change
}

// isAscendingOrder returns true if the source note's path is before the target one.
func isAscendingOrder(source, target Path) bool {
	switch {
	case source.Dir < target.Dir:
		return true
	case source.Dir > target.Dir:
		return false
	default:
		return source.Filename < target.Filename
	}
}
