package ldcache

type File struct {
	Header  *Header
	Entries []*Entry
}

func New() *File {
	h := &Header{}
	copy(h.Magic[:], magicPrefix)
	copy(h.Version[:], magicVersion)
	return &File{Header: h}
}
