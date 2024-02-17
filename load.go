package ldcache

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

var (
	MaxHeaderSize int64 = 0x1000000
)

type File struct {
	Header  *Header
	Entries []*Entry
}

func Load(in io.Reader) (*File, error) {
	// first load header
	var order binary.ByteOrder
	order = binary.BigEndian // default
	h, err := loadHeader(order, in)
	if err != nil {
		return nil, err
	}
	if h.NLibs > 0x1000000 {
		// this is too many libs, probably not using the right endian, let's flip it
		order = binary.LittleEndian
		h.flipEndian()
	}

	f := &File{
		Header: h,
	}

	// load libs
	for i := uint32(0); i < h.NLibs; i++ {
		e, err := loadEntry(order, in)
		if err != nil {
			return nil, err
		}
		f.Entries = append(f.Entries, e)
	}

	// calculate the string table offset (header size + entry size * num of entries)
	offset := uint32(headerLength) + uint32(entryLength)*h.NLibs

	// load strings table
	table := make([]byte, h.TableSize)
	_, err = io.ReadFull(in, table)
	if err != nil {
		return nil, err
	}
	// fill values in each entry based on what's found in the string table
	for _, e := range f.Entries {
		// keyPos
		e.Key, err = readFromTable(table, offset, e.keyPos)
		if err != nil {
			return nil, err
		}
		e.Value, err = readFromTable(table, offset, e.valuePos)
		if err != nil {
			return nil, err
		}
	}

	// ignore extensions for now, it's not needed for making things work
	return f, nil
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