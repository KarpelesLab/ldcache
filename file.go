package ldcache

type File struct {
	Header  *Header
	Entries []*Entry
}
