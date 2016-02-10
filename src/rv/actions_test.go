package main

import (
	"bufio"
	"bytes"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/codegangsta/cli"
)

func NewContext(args string) *cli.Context {
	set := flag.NewFlagSet("test", 0)
	set.Parse([]string{args})

	return cli.NewContext(nil, set, nil)
}

func StubOutput() (*bufio.Writer, *bytes.Buffer) {
	output := &bytes.Buffer{}
	writer := bufio.NewWriter(output)
	stdout = writer

	return writer, output
}

func Setup() {
	list := NodeList{
		"b-node.local": "127.0.0.1",
		"a-node.local": "127.0.0.2",
	}

	cacheList(list)
}

func TearDown() {
	os.Remove("/tmp/rv-cache")
}

func Test_List(t *testing.T) {
	Setup()
	writer, output := StubOutput()

	List(NewContext(""))

	writer.Flush()

	actual := output.String()

	bIndex := strings.Index(actual, "b-node.local")
	aIndex := strings.Index(actual, "a-node.local")

	if bIndex == -1 || aIndex == -1 {
		t.Errorf("Did not return both nodes. Actual: %s", actual)
	} else if aIndex > bIndex {
		t.Errorf("Nodes not sorted. Actual: %s", actual)
	}
}

func Test_CMD(t *testing.T) {
	Setup()
	writer, output := StubOutput()

	CMD(NewContext("echo b-node.local"))

	writer.Flush()

	actual := output.String()
	if !strings.Contains(actual, "127.0.0.1") {
		t.Errorf("Got %s", actual)
	}
}

func Test_NodeIP(t *testing.T) {
	Setup()
	writer, output := StubOutput()

	NodeIP(NewContext("a-node.local"))
	writer.Flush()

	actual := output.String()
	if !strings.Contains(actual, "127.0.0.2") {
		t.Errorf("Got %s", actual)
	}
}
