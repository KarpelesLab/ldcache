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

## Generating a ld.so.cache file

When generating a ld.so.cache file you need to be careful to ensure the data is properly sorted.

```go
sort.Sort(ldso.Entries)
err := ldso.SaveAs("ld.so.cache")
// handle err
```

Sorting might not be needed if you did not modify the content of the file or only removed entries without altering the overall order, but when adding entries or merging files, it will be required for the linker to be able to use the file.
