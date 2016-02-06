package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/codegangsta/cli"
)

type NodeList map[string]string

func main() {
	app := cli.NewApp()
	app.Name = "rv"
	app.Usage = "node resolver"

	app.Commands = []cli.Command{
		{
			Name:    "ip",
			Aliases: []string{"i"},
			Action:  NodeIP,
		},
		{
			Name:    "list",
			Aliases: []string{"l"},
			Action:  List,
		},
		{
			Name:    "cmd",
			Aliases: []string{"c"},
			Action:  CMD,
		},
	}

	app.Run(os.Args)
}

func CMD(c *cli.Context) {
	args := c.Args().First()

	hosts := []string{}

	for name, ip := range allNodes() {
		hosts = append(hosts, name, ip)
	}

	r := strings.NewReplacer(hosts...)
	argsWithIPs := fmt.Sprintf(r.Replace(args))

	parts := strings.Split(argsWithIPs, " ")

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	if err != nil {
		panic(err)
	}
}

func List(c *cli.Context) {
	nodes := allNodes()

	writer := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "Name\tPrivate IP Address")

	for name, ip := range nodes {
		fmt.Fprintf(writer, "%s\t%s\n", name, ip)
	}

	writer.Flush()
}

func allNodes() NodeList {
	client := ec2.New(session.New())
	params := &ec2.DescribeInstancesInput{}
	resp, err := client.DescribeInstances(params)

	if err != nil {
		panic(err)
	}

	list := NodeList{}

	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			var name string
			if instance.PrivateIpAddress == nil {
				continue
			}
			ip := *instance.PrivateIpAddress

			for _, tag := range instance.Tags {
				if *tag.Key == "Name" {
					name = *tag.Value
				}
			}

			if name != "" {
				list[name] = ip
			} else {
				list[*instance.InstanceId] = ip
			}
		}
	}

	return list
}

func NodeIP(c *cli.Context) {
	nodes := allNodes()
	node := c.Args().First()
	ip := nodes[node]

	if node == "" {
		msg := fmt.Sprintf("Node with name %s was not found", node)
		panic(msg)
	}

	fmt.Println(ip)
}
