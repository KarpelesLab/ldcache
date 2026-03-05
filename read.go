package ldcache

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

// Open opens the provided filename and will read it as a ld.so.cache file, returning
// an instance of [File] on successful read. If the file contains an old-format cache
// header (ld.so-1.7.0) followed by the new format, Open will automatically skip to
// the new-format section.
func Open(filename string) (*File, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	start := 0
	if !bytes.HasPrefix(data, []byte(magicPrefix)) {
		// File doesn't start with new format header; search for it
		// (may be preceded by old format "ld.so-1.7.0" data)
		idx := bytes.Index(data, []byte(magicPrefix))
		if idx < 0 {
			return nil, errors.New("could not find new format cache header in file")
		}
		start = idx
	}

	return Read(bytes.NewReader(data[start:]))
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
