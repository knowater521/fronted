package fronted

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCaching(t *testing.T) {
	dir, err := ioutil.TempDir("", "direct_test")
	if !assert.NoError(t, err, "Unable to create temp dir") {
		return
	}
	defer os.RemoveAll(dir)
	cacheFile := filepath.Join(dir, "cachefile.1")

	makeDirect := func() *direct {
		d := &direct{
			candidates:          make(chan masquerade, 1000),
			masquerades:         make(chan masquerade, 1000),
			maxAllowedCachedAge: 250 * time.Millisecond,
			maxCacheSize:        2,
			cacheSaveInterval:   50 * time.Millisecond,
			toCache:             make(chan masquerade, 1000),
		}
		go d.fillCache(make([]masquerade, 0), cacheFile)
		return d
	}

	now := time.Now()
	ma := masquerade{Masquerade{Domain: "a", IpAddress: "1"}, now}
	mb := masquerade{Masquerade{Domain: "b", IpAddress: "2"}, now}
	mc := masquerade{Masquerade{Domain: "c", IpAddress: "3"}, now}

	d := makeDirect()
	d.toCache <- ma
	d.toCache <- mb
	d.toCache <- mc

	readMasquerades := func() []masquerade {
		var result []masquerade
		for {
			select {
			case m := <-d.masquerades:
				result = append(result, m)
			default:
				return result
			}
		}
	}

	// Fill the cache
	time.Sleep(d.cacheSaveInterval * 2)
	d.closeCache()

	time.Sleep(50 * time.Millisecond)

	// Reopen cache file and make sure right data was in there
	d = makeDirect()
	d.prepopulateMasquerades(cacheFile)
	masquerades := readMasquerades()
	assert.Len(t, masquerades, 2, "Wrong number of masquerades read")
	assert.Equal(t, "b", masquerades[0].Domain, "Wrong masquerade at position 0")
	assert.Equal(t, "2", masquerades[0].IpAddress, "Masquerade at position 0 has wrong IpAddress")
	assert.Equal(t, "c", masquerades[1].Domain, "Wrong masquerade at position 0")
	assert.Equal(t, "3", masquerades[1].IpAddress, "Masquerade at position 1 has wrong IpAddress")
	d.closeCache()

	time.Sleep(d.maxAllowedCachedAge)
	d = makeDirect()
	d.prepopulateMasquerades(cacheFile)
	assert.Empty(t, readMasquerades(), "Cache should be empty after masquerades expire")
	d.closeCache()
}