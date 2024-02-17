package ldcache

import "encoding/binary"

type File struct {
	Header  *Header
	Entries []*Entry
	Order   binary.ByteOrder
}

func New() *File {
	h := &Header{}
	copy(h.Magic[:], magicPrefix)
	copy(h.Version[:], magicVersion)
	return &File{Header: h, Order: binary.NativeEndian}
}
