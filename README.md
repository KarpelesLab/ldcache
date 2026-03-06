[![GoDoc](https://godoc.org/github.com/KarpelesLab/ldcache?status.svg)](https://godoc.org/github.com/KarpelesLab/ldcache)
[![Go Report Card](https://goreportcard.com/badge/github.com/KarpelesLab/ldcache)](https://goreportcard.com/report/github.com/KarpelesLab/ldcache)
[![Coverage Status](https://coveralls.io/repos/github/KarpelesLab/ldcache/badge.svg?branch=master)](https://coveralls.io/github/KarpelesLab/ldcache?branch=master)

# ldcache

A Go library for reading, writing, and manipulating glibc `ld.so.cache` files (the `glibc-ld.so.cache1.1` format).

The `ld.so.cache` file is used by the dynamic linker (`ld.so`) to quickly locate shared libraries without searching the filesystem. This package can parse existing cache files, modify their entries, and generate new cache files compatible with glibc's dynamic linker.

Byte order is detected automatically when reading.

## Installation

```
go get github.com/KarpelesLab/ldcache
```

## Usage

### Opening an ld.so.cache file

```go
ldso, err := ldcache.Open("/etc/ld.so.cache")
if err != nil {
    log.Fatal(err)
}

for _, entry := range ldso.Entries {
    fmt.Println(entry)
}
```

### Creating a new cache from scratch

```go
ldso := ldcache.New()

ldso.Entries = append(ldso.Entries, &ldcache.Entry{
    Flags: 0x0303, // libc6,x86-64
    Key:   "libfoo.so.1",
    Value: "/usr/lib/libfoo.so.1.2.3",
})

sort.Sort(ldso.Entries)
err := ldso.SaveAs("/etc/ld.so.cache")
```

### Modifying and re-saving

After loading an `ld.so.cache` file, you can manipulate its `Entries` (add, remove, or filter), then generate a new file.

```go
ldso, err := ldcache.Open("/etc/ld.so.cache")
if err != nil {
    log.Fatal(err)
}

// Remove duplicates
ldso.Unique()

// Sort entries (required for the dynamic linker's binary search)
sort.Sort(ldso.Entries)

// Write to a new file
err = ldso.SaveAs("ld.so.cache")
```

### Sorting

Entries **must** be sorted before writing if you added or reordered entries.

The dynamic linker (`ld.so`) uses binary search to look up libraries in the cache: it reads the entry in the middle, compares the name with what it's looking for, then continues into the first or second half accordingly. This narrows down the result in only a few reads instead of scanning the entire file. An unsorted cache will cause the binary search to miss entries, making libraries invisible to the linker even though they are present in the file.

The sort uses glibc's `_dl_cache_libcmp` algorithm, which compares numeric segments by value rather than lexicographically (e.g. `.so.9` sorts before `.so.10`).

```go
sort.Sort(ldso.Entries)
```

Sorting may be skipped if you only removed entries without changing the relative order of the remaining ones.
