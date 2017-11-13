package main

import (
	"strings"

	"github.com/urfave/cli"

	"fmt"

	api "github.com/hashicorp/vault/api"
)

type RDSCredentials struct {
	Username string
	Password string
	Host     string
	Database string
}

func (c *RDSCredentials) ConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s/%s", c.Username, c.Password, c.Host, c.Database)
}

func (c *RDSCredentials) ConnectionEnvironment() string {
	return fmt.Sprintf("DB_USERNAME=%s\nDB_PASSWORD=%s\nDB_HOST=%s\nDB_NAME=%s\n", c.Username, c.Password, c.Host, c.Database)
}

func RegisterRDSCommands(app *cli.App) {

	app.Commands = append(app.Commands, cli.Command{
		Name:     "rds",
		Usage:    "AWS RDS",
		Category: "RDS",
		Subcommands: []cli.Command{
			{
				Name: "list",
				Action: func(c *cli.Context) error {
					env := c.String("env")
					return printDatabaseList(env)
				},
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "env",
						Usage: "the environment the database lives in (production or staging)",
						Value: "production",
					},
				},
			},
			{
				Name: "login-url",
				Action: func(c *cli.Context) error {
					env := c.String("env")
					db := c.String("db")
					entry, err := getDatabaseCredentials(env, db)
					if err != nil {
						return err
					}
					fmt.Println(entry.ConnectionString())
					return nil
				},
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "db",
						Usage: "the database to get credentials for (from 'rds list')",
					},
					cli.StringFlag{
						Name:  "env",
						Usage: "the environment the database lives in",
						Value: "production",
					},
				},
			},
			{
				Name: "login-env",
				Action: func(c *cli.Context) error {
					env := c.String("env")
					db := c.String("db")
					entry, err := getDatabaseCredentials(env, db)
					if err != nil {
						return err
					}
					fmt.Println(entry.ConnectionEnvironment())
					return nil
				},
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "db",
						Usage: "the database to get credentials for (from 'rds list')",
					},
					cli.StringFlag{
						Name:  "env",
						Usage: "the environment the database lives in",
						Value: "production",
					},
				},
			},
		},
	})
}

func getBaseVaultClient() (*api.Client, error) {
	c, err := APIClient()
	if err != nil {
		return nil, err
	}

	token := loadBaseVaultToken()
	c.SetToken(token.Token)
	return c, nil
}

func getRoleVaultClient(iamRole string, vaultRole string) (*api.Client, error) {
	c, err := APIClient()
	if err != nil {
		return nil, err
	}

	token := getCachedVaultToken(vaultRole)

	if token == nil {
		creds, err := getAssumeRoleCredentials(iamRole)
		if err != nil {
			return nil, err
		}
		token2, err := getVaultToken(c, creds, vaultRole)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		token = token2
		saveCachedVaultToken(vaultRole, token)
	}

	c.SetToken(token.Token)
	return c, nil
}

func getDatabaseCredentials(env string, db string) (*RDSCredentials, error) {
	entries := readCachedVaultRDSList()

	entry, ok := entries[db]
	if !ok {
		return nil, fmt.Errorf("DB %s not found", db)
	}

	c, err := getRoleVaultClient(entry.IamRole, entry.VaultRole)
	if err != nil {
		return nil, err
	}

	secret, err := c.Logical().Read(entry.Creds)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if secret != nil {
		return &RDSCredentials{Username: secret.Data["username"].(string),
			Password: secret.Data["password"].(string),
			Host:     entry.Host,
			Database: entry.DbName}, nil
	}

	return nil, fmt.Errorf("Unknown error getdatabasecredentials")

}

func getDatabaseList(env string) []string {
	entries := readCachedVaultRDSList()
	rdsEnvironment := "rds_staging"

	var s []string

	if env == "production" {
		rdsEnvironment = "rds_production"
	}

	for k, v := range entries {
		if strings.HasPrefix(v.Creds, rdsEnvironment) {
			s = append(s, k)
		}
	}
	return s
}

func printDatabaseList(env string) error {
	dbs := getDatabaseList(env)
	for _, v := range dbs {
		fmt.Printf("%v\n", v)
	}
	return nil
}
