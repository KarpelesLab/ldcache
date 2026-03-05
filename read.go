package ldcache

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	// oldMagic is the magic string for the legacy ld.so-1.7.0 cache format.
	oldMagic = "ld.so-1.7.0"
	// oldHeaderSize is the C sizeof(struct cache_file) from glibc: magic[11] + 1 byte
	// padding + nlibs(uint32) = 16 bytes.
	oldHeaderSize = 16
	// oldEntrySize is the C sizeof(struct file_entry): flags(int32) + key(uint32) +
	// value(uint32) = 12 bytes.
	oldEntrySize = 12
)

// Open opens the provided filename and will read it as a ld.so.cache file, returning
// an instance of [File] on successful read. If the file contains an old-format cache
// header (ld.so-1.7.0) followed by the new format, Open will parse the old header
// and skip to the new-format section at the computed offset.
func Open(filename string) (*File, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// New-format-only cache: header starts at byte 0
	if bytes.HasPrefix(data, []byte(magicPrefix)) {
		return Read(bytes.NewReader(data))
	}

	// Combined old+new format: parse old header to find new format offset.
	// See glibc elf/dl-cache.c: the new format starts at
	// ALIGN_CACHE(sizeof(struct cache_file) + nlibs * sizeof(struct file_entry))
	if !bytes.HasPrefix(data, []byte(oldMagic)) {
		return nil, errors.New("unrecognized cache file format")
	}
	if len(data) < oldHeaderSize {
		return nil, errors.New("cache file too small")
	}

	// nlibs is at offset 12 in native byte order (struct padding after char[11])
	nlibs := binary.NativeEndian.Uint32(data[12:16])
	offset := oldHeaderSize + int(nlibs)*oldEntrySize

	if offset < 0 || offset+len(magicPrefix) > len(data) {
		return nil, errors.New("cache file too small for new format section")
	}
	if !bytes.HasPrefix(data[offset:], []byte(magicPrefix)) {
		return nil, fmt.Errorf("new format header not found at expected offset %d", offset)
	}

	return Read(bytes.NewReader(data[offset:]))
}

// Read reads a ld.so.cache file from a given reader, and returns an instance of File
// on successful read. Byte order of the read file will be detected automatically
func Read(in io.Reader) (*File, error) {
	// first load header
	var order binary.ByteOrder
	order = binary.BigEndian // default
	h, err := readHeader(order, in)
	if err != nil {
		return nil, err
	}
	if h.NLibs >= 0x1000000 {
		// this is too many libs, probably not using the right endian, let's flip it
		order = binary.LittleEndian
		h.flipEndian()
	}

	f := &File{
		Header: h,
		Order:  order,
	}

	return f, f.ReadFrom(in)
}

// ReadWithOrder reads data from a given ld.so.cache file using the specified byte order
// note that if the byte order is wrong, this may result in attempts to allocate large
// amounts of memory
func ReadWithOrder(in io.Reader, order binary.ByteOrder) (*File, error) {
	// first load header
	h, err := readHeader(order, in)
	if err != nil {
		return nil, err
	}

	f := &File{
		Header: h,
		Order:  order,
	}

	return f, f.ReadFrom(in)
}

// ReadFrom will read data from a given reader, assuming the header has already been read
func (f *File) ReadFrom(in io.Reader) error {
	f.Entries = make([]*Entry, f.Header.NLibs)
	var err error

	// load libs
	for i := uint32(0); i < f.Header.NLibs; i++ {
		f.Entries[i], err = readEntry(f.Order, in)
		if err != nil {
			return err
		}
	}

	// calculate the string table offset (header size + entry size * num of entries)
	offset := uint32(headerLength) + uint32(entryLength)*f.Header.NLibs

	// load strings table
	table := make([]byte, f.Header.TableSize)
	_, err = io.ReadFull(in, table)
	if err != nil {
		return err
	}
	// fill values in each entry based on what's found in the string table
	for _, e := range f.Entries {
		// keyPos
		e.Key, err = readFromTable(table, offset, e.keyPos)
		if err != nil {
			return err
		}
		e.Value, err = readFromTable(table, offset, e.valuePos)
		if err != nil {
			return err
		}
	}

	// ignore extensions for now, it's not needed for making things work
	return nil
}

func readFromTable(table []byte, offset uint32, pos uint32) (string, error) {
	if pos < offset {
		return "", errors.New("invalid offset, too low")
	}
	pos -= offset
	if int(pos) >= len(table) {
		return "", errors.New("invalid offset, too high")
	}
	t := table[pos:]
	p := bytes.IndexByte(t, 0)
	if p != -1 {
		return string(t[:p]), nil
	}
	return string(t), nil
}
