package ldcache

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

// SaveAs creates a file with the given filename and writes the ld.so.cache data to it
func (f *File) SaveAs(filename string) error {
	fp, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fp.Close()

	return f.WriteTo(fp)
}

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
	// use a bufio writer so we don't really have to care about handling the errors.
	// bufio Writer says flush will return the latest write error if any occurs and it won't
	// process any further writes.
	wr := bufio.NewWriter(w)

	wr.Write(f.Header.Bytes(f.Order))
	for _, e := range f.Entries {
		wr.Write(e.Bytes(f.Order))
	}
	wr.Write(stringTable.Bytes())
	return wr.Flush()
}
