package ldcache_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/KarpelesLab/ldcache"
)

// test opening ldso file
func TestReadWrite(t *testing.T) {
	f, err := ldcache.Open("/pkg/main/media-video.ffmpeg.libs/.ld.so.cache")
	if err != nil {
		t.Skipf("failed to open file: %s", err)
		return
	}

	err = f.SaveAs("output.so.cache")
	if err != nil {
		t.Errorf("error: %s", err)
	}
	defer os.Remove("output.so.cache")

	// attempt to load generated file with ldconfig to see if it runs
	// ldconfig -C output.so.cache -p
	cmd := exec.Command("ldconfig", "-C", "output.so.cache", "-p")
	err = cmd.Run()
	if err != nil {
		t.Errorf("error: %s", err)
	}
}
