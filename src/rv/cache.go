package main

import (
	"bufio"
	"encoding/gob"
	"os"
	"time"

	"github.com/codegangsta/cli"
)

func cachedList() NodeList {
	fs, err := os.Stat("/tmp/rv-cache")
	if err != nil {
		return nil
	}

	if time.Now().Sub(fs.ModTime()) < CACHE_TTL {
		f, err := os.Open("/tmp/rv-cache")
		if err != nil {
			return nil
		}

		r := bufio.NewReader(f)
		dec := gob.NewDecoder(r)

		var unmarshalled NodeList
		dec.Decode(&unmarshalled)

		return unmarshalled
	} else {
		os.Remove("/tmp/rv-cache")
	}

	return nil
}

func cacheList(list NodeList) {
	f, err := os.Create("/tmp/rv-cache")
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
		os.Remove("/tmp/rv-cache")
	}
}
