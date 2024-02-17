package ldcache

import (
	"encoding/binary"
	"io"
	"unsafe"
)

var entryLength = unsafe.Sizeof(rawEntry{})

type Entry struct {
	Flags      int32
	Key, Value string
	OSVersion  uint32
	HWCap      uint64
	keyPos     uint32
	valuePos   uint32
}

type rawEntry struct {
	Flags      int32
	Key, Value uint32
	OSVersion  uint32
	HWCap      uint64
}

func loadEntry(order binary.ByteOrder, in io.Reader) (*Entry, error) {
	// load a single entry from a stream
	e := &rawEntry{}
	err := binary.Read(in, order, e)
	if err != nil {
		return nil, err
	}
	entry := &Entry{
		Flags:     e.Flags,
		keyPos:    e.Key,
		valuePos:  e.Value,
		OSVersion: e.OSVersion,
		HWCap:     e.HWCap,
	}
	return entry, nil
}
