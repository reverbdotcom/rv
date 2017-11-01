package main

import (
	"github.com/urfave/cli"

	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

func RegisterIAMCommands(app *cli.App) {

	app.Commands = append(app.Commands, cli.Command{
		Name:		"iam",
		Usage:		"AWS IAM",
		Category:	"AWS",
		Subcommands: []cli.Command{
			{
				Name: "whoami",
				Action: func(c *cli.Context) error {
					name, err := getIAMUserViaGetCallerIdentity()
					if err == nil {
						fmt.Println(name)
					}
					return err
				},
			},
			{
                Name: "ar",
                Action: func(c *cli.Context) error {
					role := c.String("role")
					creds, err := AssumeRoleCredentials(role)
					if err != nil {
						return err
					}
					fmt.Println(creds)
					return nil
                },
				Flags: []cli.Flag{
					cli.StringFlag{
						Name: "role",
						Usage: "the name of the role to assume",
					},
				},
			},
		},
	})
}

func NewSTSService(sc *sts.Credentials) (*sts.STS, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials: credentials.NewStaticCredentials(*sc.AccessKeyId, *sc.SecretAccessKey, *sc.SessionToken),
		},
	})

	if err != nil {
		return nil, err
	}

	return sts.New(sess), nil
}

func AssumeRoleCredentials(role string) (*sts.Credentials, error) {
	name, _ := getIAMUserViaGetCallerIdentity()

	params := &sts.AssumeRoleInput{
		RoleArn: aws.String(fmt.Sprintf("arn:aws:iam::590710864528:role/%s", role)),
		RoleSessionName: aws.String(name),
		DurationSeconds: aws.Int64(3600),
	}

	sc, err := getSessionCredentials()

	if err != nil {
		return nil, err
	}

	svc, err := NewSTSService(sc)

	if err != nil {
		return nil, err
	}

	resp, err := svc.AssumeRole(params)

	if err != nil {
		return nil, err
	}

	return resp.Credentials,nil 
}

func getSessionCredentials() (*sts.Credentials, error) {
	svc := sts.New(session.New())

	mfaArn, err := getMFAARN()

	if (err != nil) {
		return nil, err
	}
	
	token := &sts.GetSessionTokenInput {
		DurationSeconds: aws.Int64(3600),
		SerialNumber: aws.String(mfaArn),
		TokenCode: aws.String(getMFAToken()),
	}
	
	tokenOut, err := svc.GetSessionToken(token)

	if (err != nil) {
		return nil, err
	}

	return tokenOut.Credentials, nil
}

func getIAMUserViaGetCallerIdentity() (string, error) {
        svc := sts.New(session.New())
        input := &sts.GetCallerIdentityInput{}

        result, err := svc.GetCallerIdentity(input)
        if err != nil {
                return "", err
        } else {
                pieces := strings.Split(*result.Arn, "/")
                return pieces[len(pieces)-1], nil
        }
}

func getMFAARN() (string, error) {
	username, err := getIAMUserViaGetCallerIdentity()
	if err != nil {
		return "", nil
	}

	return fmt.Sprintf("arn:aws:iam::590710864528:mfa/%s", username), nil
}

func getMFAToken() string {
	os.Stderr.WriteString("Enter MFA code: ")

	var token string
	
	fmt.Scanln(&token)

	return token
}
