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
		"my-node.local": "127.0.0.1",
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

	if unmarshalled["my-node.local"] != "127.0.0.1" {
		t.Errorf("Expected 127.0.0.1, but got %s", unmarshalled["my-node.local"])
	}

	os.Remove("/tmp/rv-cache")
}

func Test_ReadsCacheIfUnderTTL(t *testing.T) {
	l := NodeList{
		"my-node.local": "127.0.0.1",
	}

	cacheList(l)
	list := cachedList()

	if list["my-node.local"] != "127.0.0.1" {
		t.Errorf("Expected 127.0.0.1, but got %s", list["my-node.local"])
	}

	os.Remove("/tmp/rv-cache")
}

func Test_ReturnsNothingIfCacheIsExpired(t *testing.T) {
	l := NodeList{
		"my-node.local": "127.0.0.1",
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
