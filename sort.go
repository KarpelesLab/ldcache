package ldcache

import "strings"

// EntryList implements the required methods to make Entries sortable
type EntryList []*Entry

func (e EntryList) Len() int {
	return len(e)
}

func (e EntryList) Less(i, j int) bool {
	return strings.Compare(e[i].Key, e[j].Key) > 0
}

func (e EntryList) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
