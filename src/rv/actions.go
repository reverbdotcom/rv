package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
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


type Node struct {
  id   string
  ip   string
  name string
}

type NodeList []Node

func CMD(c *cli.Context) {
	checkCache(c)

	args := c.Args().First()

  var hosts []string

	for _, v := range allNodes() {
		hosts = append(hosts, v.ip)
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
	fmt.Fprintln(writer, "Name\tInstance ID\tPrivate IP Address")

	for _, v := range nodes {
		fmt.Fprintf(writer, "%s\t%s\t%s\n", v.name, v.id, v.ip)
	}

	writer.Flush()
}

func NodeIP(c *cli.Context) {
	checkCache(c)

	nodes := allNodes()
	node := c.Args().First()

  for _,v := range nodes {
    if v.ip == node {
	    fmt.Fprintln(stdout, v.ip)
    } else if v.ip == "" {
		  fmt.Printf("Node with name %s was not found\n", node)
		  os.Exit(1)
    }
  }

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
			id := instanceId(instance)
			ip := ipAddr(instance)
			name := instanceName(instance)
      list = append(list, Node{id, ip, name})
		}
	}

	cacheList(list)
	return list
}

func instanceId(instance *ec2.Instance) string {
	if instance.InstanceId == nil {
		return "UNASSIGNED"
	} else {
		return *instance.InstanceId
	}
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
