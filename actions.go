package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/urfave/cli"
)

var stdout io.Writer = os.Stdout
var stderr io.Writer = os.Stderr

type Node struct {
	Name string
	ID   string
	IP   string
}

func CMD(c *cli.Context) {
	checkCache(c)
	args := c.Args().First()

	hosts := []string{}
	list := allNodes()

	// reverse the sort to match longer names first
	// i.e. 'node-1' should be replaced before 'node'
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name > list[j].Name
	})

	for _, node := range list {
		hosts = append(hosts, node.Name, node.IP)
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

func Grep(c *cli.Context) {
	checkCache(c)
	nodes := allNodes()

	r := c.Args().First()
	matcher := regexp.MustCompile(r)

	var matches []*Node
	for _, n := range nodes {
		if matcher.MatchString(n.Name) {
			matches = append(matches, n)
		}
	}

	printAll(matches)
}

func printAll(nodes []*Node) {
	writer := tabwriter.NewWriter(stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "Name\tIP Address\tInstance ID")

	for _, node := range nodes {
		fmt.Fprintf(writer, "%s\t%s\t%s\n", node.Name, node.IP, node.ID)
	}

	writer.Flush()
}

func List(c *cli.Context) {
	checkCache(c)
	nodes := allNodes()
	printAll(nodes)
}

func NodeIP(c *cli.Context) {
	checkCache(c)

	nodes := allNodes()
	nodeName := c.Args().First()

	for _, node := range nodes {
		if node.Name == nodeName {
			fmt.Fprintln(stdout, node.IP)
			return
		}
	}

	fmt.Printf("Node with name %s was not found\n", nodeName)
	os.Exit(1)
}

func allNodes() []*Node {
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

	var list []*Node

	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			ip := ipAddr(instance)
			name := instanceName(instance)
			node := &Node{
				ID:   *instance.InstanceId,
				Name: name,
				IP:   ip,
			}

			list = append(list, node)
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Name <= list[j].Name
	})

	uniqueNames(list)
	cacheList(list)
	return list
}

func uniqueNames(list []*Node) {
	ctr := 1
	var lastName string

	for _, node := range list {
		if node.Name == lastName {
			lastName = node.Name
			node.Name = fmt.Sprintf("%s-%d", node.Name, ctr)
			ctr++
		} else {
			lastName = node.Name
			ctr = 1
		}
	}
}

func ipAddr(instance *ec2.Instance) string {
	if instance.PrivateIpAddress == nil {
		return "UNASSIGNED"
	}

	return *instance.PrivateIpAddress

}

func instanceName(instance *ec2.Instance) string {
	var name string

	for _, tag := range instance.Tags {
		if *tag.Key == "Name" {
			name = *tag.Value
		}
	}

	if name == "" {
		return *instance.InstanceId
	}

	return name
}
