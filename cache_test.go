package main

import (
	"bufio"
	"encoding/gob"
	"os"
	"testing"
	"time"
)

func Test_WritesTheNodeList(t *testing.T) {
	list := []*Node{
		&Node{
			Name: "my-node.local",
			IP:   "127.0.0.1",
		},
	}

	cacheList(list)

	f, err := os.Open(CachePath)
	if err != nil {
		t.Errorf("Error opening file")
	}

	r := bufio.NewReader(f)
	dec := gob.NewDecoder(r)

	var unmarshalled []*Node
	dec.Decode(&unmarshalled)

	if unmarshalled[0].IP != "127.0.0.1" {
		t.Errorf("Expected 127.0.0.1, but got %s", unmarshalled[0].IP)
	}

	os.Remove(CachePath)
}

func Test_ReadsCacheIfUnderTTL(t *testing.T) {
	cacheList(
		[]*Node{
			&Node{
				Name: "my-node.local",
				IP:   "127.0.0.1",
			},
		},
	)

	list := cachedList()

	if list[0].IP != "127.0.0.1" {
		t.Errorf("Expected 127.0.0.1, but got %s", list[0].IP)
	}

	os.Remove(CachePath)
}

func Test_ReturnsNothingIfCacheIsExpired(t *testing.T) {
	cacheList(
		[]*Node{
			&Node{
				Name: "my-node.local",
				IP:   "127.0.0.1",
			},
		},
	)

	later := time.Now().Add(-61 * time.Second)
	os.Chtimes(CachePath, later, later)

	list := cachedList()

	if list != nil {
		t.Errorf("Got cache, woops")
	}

	os.Remove(CachePath)
}
