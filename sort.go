package ldcache

// EntryList implements the required methods to make Entries sortable
type EntryList []*Entry

func (e EntryList) Len() int {
	return len(e)
}

func (e EntryList) Less(i, j int) bool {
	return dlCacheLibcmp(e[i].Key, e[j].Key) > 0
}

func (e EntryList) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

// dlCacheLibcmp is an adaptation of _dl_cache_libcmp from glibc
func dlCacheLibcmp(p1, p2 string) int {
	i1, i2 := 0, 0
	len1, len2 := len(p1), len(p2)

	for i1 < len1 {
		c1 := p1[i1]

		if c1 >= '0' && c1 <= '9' {
			// c1 is a digit
			if i2 < len2 && p2[i2] >= '0' && p2[i2] <= '9' {
				// Both c1 and p2[i2] are digits: compare numerically
				val1 := int(c1 - '0')
				i1++
				val2 := int(p2[i2] - '0')
				i2++

				for i1 < len1 && p1[i1] >= '0' && p1[i1] <= '9' {
					val1 = val1*10 + int(p1[i1]-'0')
					i1++
				}

				for i2 < len2 && p2[i2] >= '0' && p2[i2] <= '9' {
					val2 = val2*10 + int(p2[i2]-'0')
					i2++
				}

				if val1 != val2 {
					return val1 - val2
				}
			} else {
				// p1 is digit, p2 not digit (or p2 ended)
				return 1
			}

		} else {
			// c1 is not a digit
			if i2 < len2 && p2[i2] >= '0' && p2[i2] <= '9' {
				// p2 is digit, p1 not
				return -1
			}

			// Compare characters directly
			var c2 byte
			if i2 < len2 {
				c2 = p2[i2]
			} else {
				c2 = 0 // p2 ended, treat as '\0'
			}

			if c1 != c2 {
				return int(c1) - int(c2)
			}

			i1++
			i2++
		}
	}

	// p1 ended; check what's left in p2
	if i2 < len2 {
		// p2 still has characters; treat p1 as '\0'
		return 0 - int(p2[i2])
	}

	// Both ended at the same time or are identical up to this point
	return 0
}
