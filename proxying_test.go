package fronted

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/getlantern/proxy"
	"github.com/stretchr/testify/assert"
)

const (
	humansText = "Google is built by a large team of engineers, designers, researchers, robots, and others in many different sites across the globe. It is updated continuously, and built with more tools and technologies than we can shake a stick at. If you'd like to help us out, see google.com/careers.\n"
)

func TestProxying(t *testing.T) {
	dir, err := ioutil.TempDir("", "direct_test")
	if !assert.NoError(t, err, "Unable to create temp dir") {
		return
	}
	defer os.RemoveAll(dir)
	cacheFile := filepath.Join(dir, "cachefile.3")
	ConfigureCachingForTest(t, cacheFile)

	conn, err := DialTimeout("d100fjyl3713ch.cloudfront.net", 30*time.Second, func(req *http.Request) {
		req.Header.Set("X-Lantern-Auth-Token", "pj6mWPafKzP26KZvUf7FIs24eB2ubjUKFvXktodqgUzZULhGeRUT0mwhyHb9jY2b")
	})
	if !assert.NoError(t, err) {
		return
	}
	defer conn.Close()

	req, _ := http.NewRequest(http.MethodGet, "https://www.google.com/humans.txt", nil)
	conn.(proxy.RequestAware).OnRequest(req)
	resp, err := httpTransport(conn, clientSessionCache).RoundTrip(req)
	if !assert.NoError(t, err) {
		return
	}
	conn.(proxy.ResponseAware).OnResponse(req, resp, err)
	if !assert.Equal(t, http.StatusOK, resp.StatusCode) {
		return
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, humansText, string(respBody))
}