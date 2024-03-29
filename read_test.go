package ldcache_test

import (
	"log"
	"os"
	"sort"
	"testing"

	"github.com/KarpelesLab/ldcache"
)

// test opening ldso file
func TestLoading(t *testing.T) {
	//f, err := os.Open("/etc/ld.so.cache")
	f, err := os.Open("/pkg/main/media-video.ffmpeg.libs/.ld.so.cache")
	if err != nil {
		t.Skipf("failed to open /etc/ld.so.cache: %s", err)
		return
	}
	defer f.Close()

	data, err := ldcache.Read(f)
	if err != nil {
		t.Errorf("error: %s", err)
	}

	sort.Sort(data.Entries)

	log.Printf("loaded data: %+v", data)
}
