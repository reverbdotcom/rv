package main

import (
	"os"

	"github.com/urfave/cli"

	"github.com/reverbdotcom/rv/pkg/vault"
	"github.com/reverbdotcom/rv/pkg/iam"
	"github.com/reverbdotcom/rv/pkg/rds"
)

func main() {
	app := cli.NewApp()
	app.Name = "rv"
	app.Usage = "reverb aws tool"
	app.Version = "0.0.4"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "clear-cache, c",
			Usage: "ensure rv cache is cleared",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "grep",
			Aliases: []string{"g"},
			Action:  Grep,
		},
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

	vault.RegisterCommands(app)
	iam.RegisterCommands(app)
	rds.RegisterCommands(app)

	app.Run(os.Args)
}
