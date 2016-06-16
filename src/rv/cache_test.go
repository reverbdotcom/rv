package main

import (
	"bufio"
	"encoding/gob"
	"os"
	"testing"
	"time"
)

func Test_WritesTheNodeList(t *testing.T) {
	list := NodeList{
    Node{
      id: "abc123",
      ip: "127.0.0.1",
      name: "nodeA",
    },
	}

	cacheList(list)

	f, err := os.Open("/tmp/rv-cache")
	if err != nil {
		t.Errorf("Error opening file")
	}

	r := bufio.NewReader(f)
	dec := gob.NewDecoder(r)

	var unmarshalled NodeList
	dec.Decode(&unmarshalled)

	if unmarshalled[0].ip != "127.0.0.1" {
		t.Errorf("Expected 127.0.0.1, but got %s", unmarshalled[0])
	}

	os.Remove("/tmp/rv-cache")
}

func Test_ReadsCacheIfUnderTTL(t *testing.T) {
	l := NodeList{
    Node{
      id: "abc123",
      ip: "127.0.0.1",
      name: "nodeA",
    },
	}

	cacheList(l)
	list := cachedList()

	if list[0].ip != "127.0.0.1" {
		t.Errorf("Expected 127.0.0.1, but got %s", list[0])
	}

	os.Remove("/tmp/rv-cache")
}

func Test_ReturnsNothingIfCacheIsExpired(t *testing.T) {
	l := NodeList{
    Node{
      id: "abc123",
      ip: "127.0.0.1",
      name: "nodeA",
    },
	}

	cacheList(l)

	later := time.Now().Add(-61 * time.Second)
	os.Chtimes("/tmp/rv-cache", later, later)

	list := cachedList()

	if list != nil {
		t.Errorf("Got cache, woops")
	}

	os.Remove("/tmp/rv-cache")
}
