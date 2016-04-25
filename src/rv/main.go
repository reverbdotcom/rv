package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "rv"
	app.Usage = "aws instance resolver"
	app.Version = "0.0.4"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "clear-cache, c",
			Usage: "ensure rv cache is cleared",
		},
	}

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
