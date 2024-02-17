package ldcache

import "encoding/binary"

// File represents a ld.so.cache file data
type File struct {
	Header  *Header
	Entries EntryList
	Order   binary.ByteOrder
}

// New returns a new empty ld.so.cache file, to which entries can be added
func New() *File {
	h := &Header{}
	copy(h.Magic[:], magicPrefix)
	copy(h.Version[:], magicVersion)
	return &File{Header: h, Order: binary.NativeEndian}
}

type uniqueKey struct {
	Key   string
	Flags Flags
}

// Unique checks for any duplicate vlaues in Entries and ensures there is only one
// entry for each filename + Flags
func (f *File) Unique() {
	found := make(map[uniqueKey]bool)
	newEntries := make([]*Entry, 0, len(f.Entries))

	for _, e := range f.Entries {
		k := uniqueKey{e.Key, e.Flags}
		if _, ok := found[k]; ok {
			// already there, skip
			continue
		}
		newEntries = append(newEntries, e)
		found[k] = true
	}
	f.Entries = newEntries
}
