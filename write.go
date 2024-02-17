package ldcache

import (
	"bufio"
	"bytes"
	"io"
)

// WriteTo updates information found in Header and writes the file to the given io.Writer.
func (f *File) WriteTo(w io.Writer) error {
	// generate string table
	f.Header.NLibs = uint32(len(f.Entries))
	offset := uint32(headerLength) + uint32(entryLength)*f.Header.NLibs
	stringTable := &bytes.Buffer{}

	for _, e := range f.Entries {
		e.keyPos = uint32(stringTable.Len()) + offset
		stringTable.WriteString(e.Key)
		stringTable.WriteByte(0)
		e.valuePos = uint32(stringTable.Len()) + offset
		stringTable.WriteString(e.Value)
		stringTable.WriteByte(0)
	}
	f.Header.TableSize = uint32(stringTable.Len())

	// we're ready to write
	wr := bufio.NewWriter(w)
	wr.Write(f.Header.Bytes(f.Order))
	for _, e := range f.Entries {
		wr.Write(e.Bytes(f.Order))
	}
	wr.Write(stringTable.Bytes())
	return wr.Flush()
}
