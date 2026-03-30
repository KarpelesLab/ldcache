package ldcache

import (
	"bytes"
	"io"
	"os"
	"strings"
)

// SaveAs creates a file with the given filename and writes the ld.so.cache data to it.
func (f *File) SaveAs(filename string) error {
	fp, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fp.Close()

	_, err = f.WriteTo(fp)
	if err != nil {
		return err
	}
	return fp.Close()
}

// WriteTo updates information found in Header and writes the file to the given
// [io.Writer]. It implements the [io.WriterTo] interface.
func (f *File) WriteTo(w io.Writer) (int64, error) {
	// generate string table
	f.Header.NLibs = uint32(len(f.Entries))
	offset := uint32(headerLength) + uint32(entryLength)*f.Header.NLibs
	stringTable := &bytes.Buffer{}

	for _, e := range f.Entries {
		e.valuePos = uint32(stringTable.Len()) + offset
		stringTable.WriteString(e.Value)
		stringTable.WriteByte(0)
		// Optimization matching glibc: if the key is a suffix of the value,
		// point the key into the value string instead of writing it separately.
		if strings.HasSuffix(e.Value, e.Key) {
			e.keyPos = e.valuePos + uint32(len(e.Value)-len(e.Key))
		} else {
			e.keyPos = uint32(stringTable.Len()) + offset
			stringTable.WriteString(e.Key)
			stringTable.WriteByte(0)
		}
	}
	f.Header.TableSize = uint32(stringTable.Len())

	// we're ready to write
	var n int64
	nn, err := w.Write(f.Header.Bytes(f.Order))
	n += int64(nn)
	if err != nil {
		return n, err
	}
	for _, e := range f.Entries {
		nn, err = w.Write(e.Bytes(f.Order))
		n += int64(nn)
		if err != nil {
			return n, err
		}
	}
	nn, err = w.Write(stringTable.Bytes())
	n += int64(nn)
	return n, err
}
