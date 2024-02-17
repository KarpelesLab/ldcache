package ldcache

import "encoding/binary"

// File represents a ld.so.cache file data
type File struct {
	Header  *Header
	Entries []*Entry
	Order   binary.ByteOrder
}

// New returns a new empty ld.so.cache file, to which entries can be added
func New() *File {
	h := &Header{}
	copy(h.Magic[:], magicPrefix)
	copy(h.Version[:], magicVersion)
	return &File{Header: h, Order: binary.NativeEndian}
}
