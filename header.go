package ldcache

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unsafe"
)

var headerLength = unsafe.Sizeof(Header{})

type Header struct {
	Magic     [len(magicPrefix)]byte
	Version   [3]byte
	NLibs     uint32
	TableSize uint32
	_         [3]uint32 // unused
	_         uint64    // 8 bytes align
}

func readHeader(order binary.ByteOrder, r io.Reader) (*Header, error) {
	h := &Header{}
	err := binary.Read(r, order, h)
	if err != nil {
		return nil, err
	}

	// check magic and version
	if !bytes.Equal(h.Magic[:], []byte(magicPrefix)) {
		return nil, errors.New("invalid header: bad magic")
	}
	if !bytes.Equal(h.Version[:], []byte(magicVersion)) {
		return nil, errors.New("invalid header: bad version")
	}

	return h, nil
}

func (h *Header) flipEndian() {
	// flip NLibs and TableSize between big and little endian by reversing the 4 bytes
	// ABCD to DCBA
	h.NLibs = (h.NLibs&0xff000000)>>24 | (h.NLibs&0xff0000)>>8 | (h.NLibs&0xff00)<<8 | (h.NLibs&0xff)<<24
	h.TableSize = (h.TableSize&0xff000000)>>24 | (h.TableSize&0xff0000)>>8 | (h.TableSize&0xff00)<<8 | (h.TableSize&0xff)<<24
}

func (h *Header) Bytes(order binary.ByteOrder) []byte {
	buf := &bytes.Buffer{}
	binary.Write(buf, order, h)
	return buf.Bytes()
}

func (h *Header) String() string {
	return fmt.Sprintf("ld.so.cache header (%s %s), %d libs", h.Magic, h.Version, h.NLibs)
}
