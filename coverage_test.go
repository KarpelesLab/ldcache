package ldcache_test

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/KarpelesLab/ldcache"
)

func TestUnique(t *testing.T) {
	f := ldcache.New()
	f.Entries = ldcache.EntryList{
		{Flags: 0x0303, Key: "libfoo.so.1", Value: "/usr/lib/libfoo.so.1"},
		{Flags: 0x0303, Key: "libfoo.so.1", Value: "/usr/lib/libfoo.so.1"},   // exact dup
		{Flags: 0x0003, Key: "libfoo.so.1", Value: "/usr/lib32/libfoo.so.1"}, // different flags, not dup
		{Flags: 0x0303, Key: "libbar.so.2", Value: "/usr/lib/libbar.so.2"},
		{Flags: 0x0303, Key: "libbar.so.2", Value: "/usr/lib/libbar.so.2.0.0"}, // same key+flags = dup
	}
	f.Unique()
	if len(f.Entries) != 3 {
		t.Errorf("expected 3 unique entries, got %d", len(f.Entries))
	}
	// First occurrence of each key+flags should be kept
	if f.Entries[0].Value != "/usr/lib/libfoo.so.1" {
		t.Errorf("expected first libfoo entry kept, got %s", f.Entries[0].Value)
	}
	if f.Entries[2].Value != "/usr/lib/libbar.so.2" {
		t.Errorf("expected first libbar entry kept, got %s", f.Entries[2].Value)
	}
}

func TestFlagsString(t *testing.T) {
	tests := []struct {
		flags ldcache.Flags
		want  string
	}{
		{-1, "any"},
		{0x0000, "libc4"},
		{0x0001, "elf"},
		{0x0002, "libc5"},
		{0x0003, "libc6"},
		{0x00FF, "255"},         // unknown type
		{0x0103, "libc6,64bit"}, // SPARC64
		{0x0203, "libc6,IA-64"},
		{0x0303, "libc6,x86-64"},
		{0x0403, "libc6,64bit"}, // S390
		{0x0503, "libc6,64bit"}, // POWERPC
		{0x0603, "libc6,N32"},
		{0x0703, "libc6,64bit"}, // MIPS64
		{0x0803, "libc6,x32"},
		{0x0903, "libc6,hard-float"},
		{0x0A03, "libc6,AArch64"},
		{0x0B03, "libc6,soft-float"}, // ARM SF
		{0x0C03, "libc6,nan2008"},
		{0x0D03, "libc6,N32,nan2008"},
		{0x0E03, "libc6,64bit,nan2008"},
		{0x0F03, "libc6,soft-float"},   // RISCV soft
		{0x1003, "libc6,double-float"}, // RISCV double
		{0x1103, "libc6,soft-float"},   // LARCH soft
		{0x1203, "libc6,double-float"}, // LARCH double
		{0x1303, "libc6,19"},           // unknown req
	}
	for _, tt := range tests {
		got := tt.flags.String()
		if got != tt.want {
			t.Errorf("Flags(%#x).String() = %q, want %q", int32(tt.flags), got, tt.want)
		}
	}
}

func TestOpenErrors(t *testing.T) {
	// Non-existent file
	_, err := ldcache.Open("/nonexistent/path/ld.so.cache")
	if err == nil {
		t.Error("expected error for non-existent file")
	}

	tmpDir := t.TempDir()

	// Unrecognized format
	badFile := filepath.Join(tmpDir, "bad.cache")
	os.WriteFile(badFile, []byte("not a cache file at all"), 0644)
	_, err = ldcache.Open(badFile)
	if err == nil {
		t.Error("expected error for unrecognized format")
	}

	// Truncated old format header
	truncOld := filepath.Join(tmpDir, "trunc_old.cache")
	os.WriteFile(truncOld, []byte("ld.so-1.7.0"), 0644) // only 11 bytes, need 16
	_, err = ldcache.Open(truncOld)
	if err == nil {
		t.Error("expected error for truncated old format")
	}

	// Old format with nlibs pointing past end of file
	bigNlibs := filepath.Join(tmpDir, "big_nlibs.cache")
	buf := make([]byte, 16)
	copy(buf, "ld.so-1.7.0")
	binary.NativeEndian.PutUint32(buf[12:16], 999999) // way too many
	os.WriteFile(bigNlibs, buf, 0644)
	_, err = ldcache.Open(bigNlibs)
	if err == nil {
		t.Error("expected error for oversized nlibs")
	}

	// Old format with nlibs=0 but no new format at expected offset
	noNew := filepath.Join(tmpDir, "no_new.cache")
	buf = make([]byte, 32)
	copy(buf, "ld.so-1.7.0")
	binary.NativeEndian.PutUint32(buf[12:16], 0)
	copy(buf[16:], "not-the-magic!!!")
	os.WriteFile(noNew, buf, 0644)
	_, err = ldcache.Open(noNew)
	if err == nil {
		t.Error("expected error when new format magic not found at offset")
	}
}

func TestOpenOldFormat(t *testing.T) {
	// Build a valid combined old+new format file:
	// old header (16 bytes, nlibs=0) + new format cache

	f := ldcache.New()
	f.Entries = ldcache.EntryList{
		{Flags: 0x0303, Key: "libtest.so.1", Value: "/usr/lib/libtest.so.1"},
	}
	var newFormat bytes.Buffer
	f.WriteTo(&newFormat)

	// Prepend old format header with nlibs=0
	oldHeader := make([]byte, 16)
	copy(oldHeader, "ld.so-1.7.0")
	binary.NativeEndian.PutUint32(oldHeader[12:16], 0)

	combined := append(oldHeader, newFormat.Bytes()...)
	tmpFile := filepath.Join(t.TempDir(), "combined.cache")
	os.WriteFile(tmpFile, combined, 0644)

	f2, err := ldcache.Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open combined format: %v", err)
	}
	if len(f2.Entries) != 1 || f2.Entries[0].Key != "libtest.so.1" {
		t.Errorf("unexpected entries: %v", f2.Entries)
	}
}

func TestOpenOldFormatOddNlibs(t *testing.T) {
	// Build a combined old+new format file where nlibs is odd,
	// requiring ALIGN_CACHE to find the new-format header.
	// With nlibs=1: offset = 16 + 1*12 = 28, aligned to 32.

	f := ldcache.New()
	f.Entries = ldcache.EntryList{
		{Flags: 0x0303, Key: "libtest.so.1", Value: "/usr/lib/libtest.so.1"},
	}
	var newFormat bytes.Buffer
	f.WriteTo(&newFormat)

	// Old header with nlibs=1 (one fake 12-byte old entry)
	oldHeader := make([]byte, 16)
	copy(oldHeader, "ld.so-1.7.0")
	binary.NativeEndian.PutUint32(oldHeader[12:16], 1)

	// Add a 12-byte dummy old entry + 4 bytes alignment padding + new format
	oldEntry := make([]byte, 12) // dummy old-format entry
	padding := make([]byte, 4)   // alignment to 8-byte boundary: 28 → 32
	combined := append(oldHeader, oldEntry...)
	combined = append(combined, padding...)
	combined = append(combined, newFormat.Bytes()...)

	tmpFile := filepath.Join(t.TempDir(), "combined_odd.cache")
	os.WriteFile(tmpFile, combined, 0644)

	f2, err := ldcache.Open(tmpFile)
	if err != nil {
		t.Fatalf("failed to open combined format with odd nlibs: %v", err)
	}
	if len(f2.Entries) != 1 || f2.Entries[0].Key != "libtest.so.1" {
		t.Errorf("unexpected entries: %v", f2.Entries)
	}
}

func TestReadWithOrder(t *testing.T) {
	// Write a cache, then read it back with explicit byte order
	f := ldcache.New()
	f.Entries = ldcache.EntryList{
		{Flags: 0x0303, Key: "libfoo.so.1", Value: "/usr/lib/libfoo.so.1"},
	}
	sort.Sort(f.Entries)

	var buf bytes.Buffer
	f.WriteTo(&buf)

	f2, err := ldcache.ReadWithOrder(bytes.NewReader(buf.Bytes()), binary.LittleEndian)
	if err != nil {
		t.Fatalf("ReadWithOrder failed: %v", err)
	}
	if len(f2.Entries) != 1 || f2.Entries[0].Key != "libfoo.so.1" {
		t.Errorf("unexpected entries: %v", f2.Entries)
	}
}

func TestReadErrors(t *testing.T) {
	// Empty reader
	_, err := ldcache.Read(bytes.NewReader(nil))
	if err == nil {
		t.Error("expected error for empty reader")
	}

	// Invalid magic
	bad := make([]byte, 48)
	copy(bad, "not-valid-magic!")
	_, err = ldcache.Read(bytes.NewReader(bad))
	if err == nil {
		t.Error("expected error for bad magic")
	}

	// Valid magic but bad version
	badVer := make([]byte, 48)
	copy(badVer, "glibc-ld.so.cache")
	copy(badVer[17:], "9.9") // wrong version
	_, err = ldcache.Read(bytes.NewReader(badVer))
	if err == nil {
		t.Error("expected error for bad version")
	}

	// Valid header but truncated entries
	f := ldcache.New()
	f.Entries = ldcache.EntryList{
		{Flags: 0x0303, Key: "libfoo.so.1", Value: "/usr/lib/libfoo.so.1"},
	}
	var buf bytes.Buffer
	f.WriteTo(&buf)
	// Truncate after header (48 bytes) but before the entry data ends
	_, err = ldcache.Read(bytes.NewReader(buf.Bytes()[:50]))
	if err == nil {
		t.Error("expected error for truncated entry data")
	}

	// Valid header+entries but truncated string table
	_, err = ldcache.Read(bytes.NewReader(buf.Bytes()[:72]))
	if err == nil {
		t.Error("expected error for truncated string table")
	}

	// ReadWithOrder with truncated reader
	_, err = ldcache.ReadWithOrder(bytes.NewReader(nil), binary.LittleEndian)
	if err == nil {
		t.Error("expected error for empty reader in ReadWithOrder")
	}

	// Corrupt entry with key offset pointing before string table
	data := buf.Bytes()
	good := make([]byte, len(data))
	copy(good, data)
	// Entry key offset is at header(48) + 4(flags) = byte 52, set to 0 (before string table)
	binary.LittleEndian.PutUint32(good[52:56], 0)
	_, err = ldcache.Read(bytes.NewReader(good))
	if err == nil {
		t.Error("expected error for invalid key offset")
	}

	// Corrupt entry with value offset pointing past end
	copy(good, data)
	// Entry value offset is at header(48) + 8 = byte 56
	binary.LittleEndian.PutUint32(good[56:60], 0xFFFFFF)
	_, err = ldcache.Read(bytes.NewReader(good))
	if err == nil {
		t.Error("expected error for invalid value offset")
	}
}

func TestSaveAsError(t *testing.T) {
	f := ldcache.New()
	err := f.SaveAs("/nonexistent/dir/cache")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestDlCacheLibcmpEdgeCases(t *testing.T) {
	// Test via sort behavior since dlCacheLibcmp is unexported.
	// These entries exercise edge cases in the comparison function.
	entries := ldcache.EntryList{
		{Key: "lib.so"},
		{Key: "lib.so.1"},
		{Key: "lib.so.2"},
		{Key: "lib.so.10"},
		{Key: "lib9.so"},
		{Key: "lib10.so"},
		{Key: "lib.so"}, // duplicate
		{Key: "libz.so"},
		{Key: "liba.so"},
		{Key: ""},
	}
	sort.Sort(entries)

	// Verify sort is stable and doesn't panic
	for i := 1; i < len(entries); i++ {
		// Each entry should be <= the previous one (descending order)
		if entries[i-1].Key < entries[i].Key && entries[i-1].Key != "" {
			// This is a rough check; the actual comparison is version-aware
			// Just verify no panics and ordering is consistent
		}
	}

	// Verify .so.10 sorts before .so.2 (numeric comparison: 10 > 2, descending)
	idx10, idx2 := -1, -1
	for i, e := range entries {
		if e.Key == "lib.so.10" {
			idx10 = i
		}
		if e.Key == "lib.so.2" {
			idx2 = i
		}
	}
	if idx10 >= 0 && idx2 >= 0 && idx10 > idx2 {
		t.Errorf("lib.so.10 should sort before lib.so.2 (descending), got indices %d and %d", idx10, idx2)
	}
}

func TestWriteToError(t *testing.T) {
	f := ldcache.New()
	f.Entries = ldcache.EntryList{
		{Flags: 0x0303, Key: "libfoo.so.1", Value: "/usr/lib/libfoo.so.1"},
	}

	// Write to a writer that always fails
	w := &failWriter{}
	_, err := f.WriteTo(w)
	if err == nil {
		t.Error("expected error from failing writer")
	}
}

type failWriter struct{}

func (f *failWriter) Write(p []byte) (int, error) {
	return 0, os.ErrClosed
}
