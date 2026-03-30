package ldcache

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"
)

var entryLength = unsafe.Sizeof(rawEntry{})

// Entry represents an entry in a ld.so.cache file
type Entry struct {
	Flags      Flags
	Key, Value string
	OSVersion  uint32
	HWCap      uint64
	keyPos     uint32
	valuePos   uint32
}

type rawEntry struct {
	Flags      Flags
	Key, Value uint32
	OSVersion  uint32
	HWCap      uint64
}

func readEntry(order binary.ByteOrder, in io.Reader) (*Entry, error) {
	// read a single entry from a stream
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

// Bytes returns the entry serialized as it would appear in an ld.so.cache file.
func (e *Entry) Bytes(order binary.ByteOrder) []byte {
	buf := make([]byte, entryLength)
	order.PutUint32(buf[:4], uint32(e.Flags))
	order.PutUint32(buf[4:8], e.keyPos)
	order.PutUint32(buf[8:12], e.valuePos)
	order.PutUint32(buf[12:16], e.OSVersion)
	order.PutUint64(buf[16:24], e.HWCap)
	return buf
}

// String returns a human-readable representation of the entry, similar to
// the output format of ldconfig -p.
func (e *Entry) String() string {
	return fmt.Sprintf("%s (%s) => %s (OSVersion=%d HWCap=%d)", e.Key, e.Flags, e.Value, e.OSVersion, e.HWCap)
}
