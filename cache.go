package main

import (
	"bufio"
	"encoding/gob"
	"os"
	"time"

	"github.com/codegangsta/cli"
)

const CachePath = "/tmp/rv-cache"
const CacheTTL = 60 * time.Second

func cachedList() []*Node {
	fs, err := os.Stat(CachePath)
	if err != nil {
		return nil
	}

	if time.Now().Sub(fs.ModTime()) < CacheTTL {
		f, err := os.Open(CachePath)
		if err != nil {
			return nil
		}

		r := bufio.NewReader(f)
		dec := gob.NewDecoder(r)

		var unmarshalled []*Node
		dec.Decode(&unmarshalled)

		return unmarshalled
	} else {
		os.Remove(CachePath)
	}

	return nil
}

func cacheList(list []*Node) {
	f, err := os.Create(CachePath)
	if err != nil {
		return
	}

	defer f.Close()

	w := bufio.NewWriter(f)
	enc := gob.NewEncoder(w)

	enc.Encode(list)
	w.Flush()
}

func checkCache(c *cli.Context) {
	clearCache := c.GlobalBool("clear-cache")
	if clearCache {
		os.Remove(CachePath)
	}
}
