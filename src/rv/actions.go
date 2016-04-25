package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/codegangsta/cli"
)

const CACHE_TTL = 60 * time.Second

var stdout io.Writer = os.Stdout
var stderr io.Writer = os.Stderr

type NodeList map[string]string

func CMD(c *cli.Context) {
	checkCache(c)

	args := c.Args().First()

	hosts := []string{}

	for name, ip := range allNodes() {
		hosts = append(hosts, name, ip)
	}

	r := strings.NewReplacer(hosts...)
	argsWithIPs := fmt.Sprintf(r.Replace(args))

	parts := strings.Split(argsWithIPs, " ")

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func List(c *cli.Context) {
	checkCache(c)

	nodes := allNodes()

	writer := tabwriter.NewWriter(stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "Name\tPrivate IP Address")

	nodeNames := sortedNodeNames(nodes)

	for _, name := range nodeNames {
		fmt.Fprintf(writer, "%s\t%s\n", name, nodes[name])
	}

	writer.Flush()
}

func NodeIP(c *cli.Context) {
	checkCache(c)

	nodes := allNodes()
	node := c.Args().First()
	ip := nodes[node]

	if ip == "" {
		fmt.Printf("Node with name %s was not found\n", node)
		os.Exit(1)
	}

	fmt.Fprintln(stdout, ip)
}

func sortedNodeNames(nodeList NodeList) []string {
	names := make([]string, len(nodeList))
	i := 0
	for name, _ := range nodeList {
		names[i] = name
		i++
	}
	sort.Strings(names)
	return names
}

func allNodes() NodeList {
	if list := cachedList(); list != nil {
		return list
	}

	client := ec2.New(session.New())
	params := &ec2.DescribeInstancesInput{}
	resp, err := client.DescribeInstances(params)

	if err != nil {
		fmt.Printf("Error connecting to AWS: %s\n", err)
		os.Exit(1)

		return nil
	}

	list := NodeList{}

	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			ip := ipAddr(instance)
			name := instanceName(instance)

			list[name] = ip
		}
	}

	cacheList(list)
	return list
}

func ipAddr(instance *ec2.Instance) string {
	if instance.PrivateIpAddress == nil {
		return "UNASSIGNED"
	} else {
		return *instance.PrivateIpAddress
	}
}

func instanceName(instance *ec2.Instance) string {
	var name string

	for _, tag := range instance.Tags {
		if *tag.Key == "Name" {
			name = *tag.Value
		}
	}

	if name != "" {
		return name
	} else {
		return *instance.InstanceId
	}
}
