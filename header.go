package ldcache

import (
	"bytes"
	"encoding/binary"
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
	return h, binary.Read(r, order, h)
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
