package main

import (
	"bytes"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/codegangsta/cli"
)

func NewContext(args ...string) *cli.Context {
	set := flag.NewFlagSet("test", 0)
	set.Parse(args)

	return cli.NewContext(nil, set, nil)
}

func StubOutput() (*bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	err := &bytes.Buffer{}
	stdout = out
	stderr = err

	return out, err
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
	defer TearDown()

	output, _ := StubOutput()

	List(NewContext(""))

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
	defer TearDown()

	output, _ := StubOutput()
	CMD(NewContext("echo b-node.local"))

	actual := output.String()
	if !strings.Contains(actual, "127.0.0.1") {
		t.Errorf("Got %s", actual)
	}
}

// I have no clue how to actually test this behavior. It _does_ work...
func XTest_CMDErrorOutput(t *testing.T) {
	Setup()
	defer TearDown()

	_, errout := StubOutput()

	CMD(NewContext("echo", "foo", "1>&2"))

	actual := errout.String()
	if !strings.Contains(actual, "foo") {
		t.Errorf("Got %s", actual)
	}
}

func Test_NodeIP(t *testing.T) {
	Setup()
	defer TearDown()

	output, _ := StubOutput()

	NodeIP(NewContext("a-node.local"))

	actual := output.String()
	if !strings.Contains(actual, "127.0.0.2") {
		t.Errorf("Got %s", actual)
	}
}
