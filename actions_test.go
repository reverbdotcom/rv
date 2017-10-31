package main

import (
	"bytes"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/urfave/cli"
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
	list := []*Node{
		&Node{
			Name: "b-node.local",
			IP:   "127.0.0.1",
		},
		&Node{
			Name: "a-node.local",
			IP:   "127.0.0.2",
		},
	}

	cacheList(list)
}

func TearDown() {
	os.Remove(CachePath)
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

func Test_Grep(t *testing.T) {
	Setup()
	defer TearDown()

	output, _ := StubOutput()

	Grep(NewContext("^a.*"))

	actual := output.String()
	if !strings.Contains(actual, "127.0.0.2") {
		t.Errorf("Got %s", actual)
	}

	if strings.Contains(actual, "127.0.0.1") {
		t.Errorf("Got %s, but expected it to be filtered out", actual)
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
