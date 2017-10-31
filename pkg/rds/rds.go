package rds

import (
	"github.com/urfave/cli"

	"fmt"

	api "github.com/hashicorp/vault/api"

	"github.com/reverbdotcom/rv/pkg/vault"
	"github.com/reverbdotcom/rv/pkg/iam"
)

func RegisterCommands(app *cli.App) {

	app.Commands = append(app.Commands, cli.Command{
		Name:		"rds",
		Usage:		"AWS RDS",
		Category:	"RDS",
		Subcommands: []cli.Command{
			{
				Name: "list",
				Action: func(c *cli.Context) error {
					return ListDatabases()
				},
			},
			{
				Name: "getcreds",
				Action: func(c *cli.Context) error {
					env := c.String("env")
					db := c.String("db")
					return GetDatabaseCredentials(env, db)
				},
				Flags: []cli.Flag{
					cli.StringFlag{
						Name: "db",
						Usage: "the database to get credentials for",
					},
					cli.StringFlag{
						Name: "env",
						Usage: "the environment the database lives in",
					},

				},
			},

		},
	})
}

func getInfraDevVaultClient() (*api.Client, error) {
	c, err := vault.ApiClient()
	if err != nil {
		return nil, err
	}

	token := vault.GetCachedVaultToken("ops-infra-dev")

	if token == "" {

		creds, err := iam.AssumeRoleCredentials("ops/infra-dev")
		if err != nil {
			return nil, err
		}

		token2, err := vault.GetVaultToken(c, creds, "ops-infra-dev")
		if err != nil {
			return nil, err
		}

		token = token2;
	}

	c.SetToken(token)
	return c, nil
}

func GetDatabaseCredentials(env string, db string) error {
	c, err := getInfraDevVaultClient()
	if err != nil {
		return err
	}

	vaultValue := env + "/creds/" + db

	secret, err := c.Logical().Read(vaultValue)
	if err != nil {
		return err
	}

	if secret != nil {
		fmt.Printf("%v\n", secret.Data)
	}

	return nil

}

func ListDatabases() error {
	c, err := getInfraDevVaultClient()
	if err != nil {
		return err
	}

	envs := []string{"rds_production", "rds_production_readonly", "rds_staging"}
	for _, env := range envs {
		secret, err := c.Logical().List(env + "/roles")
		if err != nil {
			return err
		}
	
		if secret != nil {
			fmt.Printf("%s: %v\n", env, secret.Data["keys"])
		}
	}

	return nil
}
