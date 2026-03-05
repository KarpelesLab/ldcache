// Package ldcache provides reading, writing, and manipulation of glibc
// ld.so.cache files (the "new" glibc-ld.so.cache1.1 format).
//
// The ld.so.cache file is used by the dynamic linker (ld.so) to quickly
// locate shared libraries without searching the filesystem. This package
// can parse existing cache files, modify their entries, and generate new
// cache files that are compatible with glibc's dynamic linker.
//
// Byte order is detected automatically when reading, and preserved when
// writing.
package ldcache

const (
	magicPrefix  = "glibc-ld.so.cache"
	magicVersion = "1.1"
)
