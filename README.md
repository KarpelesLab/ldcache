[![GoDoc](https://godoc.org/github.com/KarpelesLab/ldcache?status.svg)](https://godoc.org/github.com/KarpelesLab/ldcache)

# ldcache

Library to read/write ld.so.cache files.

## Opening a ld.so.cache file

```go
ldso, err := ldcache.Open("/etc/ld.so.cache")
if err != nil {
    // ...
}
```

After loading a ld.so.cache file, it is possible to manipulate its `Entries`, such as adding or removing entries, then generate a new file that will match the updated data.
