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
	app.Version = "0.0.2"

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
		os.Exit(1)
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

func NodeIP(c *cli.Context) {
	nodes := allNodes()
	node := c.Args().First()
	ip := nodes[node]

	if node == "" {
		fmt.Printf("Node with name %s was not found\n", node)
		os.Exit(1)
	}

	fmt.Println(ip)
}

func allNodes() NodeList {
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
