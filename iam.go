package main

import (
	"time"

	"github.com/urfave/cli"

	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

/*
type AwsCredentials struct {
	IamArn      string
	Credentials *sts.Credentials
}
*/
type AwsCredentials struct {
	AccessKeyId     string
	Expiration      time.Time
	SecretAccessKey string
	SessionToken    string
}

type AwsCredentialsWithCallerArn struct {
	IamArn      string
	Credentials *AwsCredentials
}

func (c AwsCredentials) expired() bool {
	return c.Expiration.Before(time.Now())
}

func RegisterIAMCommands(app *cli.App) {

	app.Commands = append(app.Commands, cli.Command{
		Name:     "iam",
		Usage:    "AWS IAM",
		Category: "AWS",
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
					creds, err := getAssumeRoleCredentials(role)
					if err != nil {
						return err
					}
					fmt.Println(creds)
					return nil
				},
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "role",
						Usage: "the name of the role to assume",
					},
				},
			},
		},
	})
}

func NewSTSService(sc *AwsCredentials) (*sts.STS, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials: credentials.NewStaticCredentials(sc.AccessKeyId, sc.SecretAccessKey, sc.SessionToken),
		},
	})

	if err != nil {
		return nil, err
	}

	return sts.New(sess), nil
}

func getAssumeRoleCredentials(role string) (*AwsCredentialsWithCallerArn, error) {

	cachedCred := getCachedAwsCredentials(role)
	if cachedCred != nil {
		return &AwsCredentialsWithCallerArn{Credentials: cachedCred, IamArn: role}, nil
	}

	name, _ := getIAMUserViaGetCallerIdentity()

	params := &sts.AssumeRoleInput{
		RoleArn:         aws.String(role),
		RoleSessionName: aws.String(name),
		DurationSeconds: aws.Int64(3600),
	}

	sc := getCachedSessionCredentials()
	if sc == nil {
		sc, _ = getSessionCredentials()
	}
	svc, err := NewSTSService(sc.Credentials)

	if err != nil {
		return nil, err
	}

	resp, err := svc.AssumeRole(params)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	awsCreds := &AwsCredentials{AccessKeyId: *resp.Credentials.AccessKeyId,
		SecretAccessKey: *resp.Credentials.SecretAccessKey,
		SessionToken:    *resp.Credentials.SessionToken,
		Expiration:      *resp.Credentials.Expiration}

	saveCachedAwsCredentials(role, awsCreds)

	return &AwsCredentialsWithCallerArn{Credentials: awsCreds, IamArn: role}, nil

}

func getCachedSessionCredentials() *AwsCredentialsWithCallerArn {
	svc := sts.New(session.New())

	// Get the MFA ARN
	input := &sts.GetCallerIdentityInput{}

	result, err := svc.GetCallerIdentity(input)
	if err != nil {
		return nil
	}

	callerIdentity := *result.Arn
	cachedCred := getCachedAwsCredentials(callerIdentity)

	fmt.Println("in get cached")

	if cachedCred != nil {
		return &AwsCredentialsWithCallerArn{Credentials: cachedCred, IamArn: callerIdentity}
	}

	return nil
}

func getSessionCredentials() (*AwsCredentialsWithCallerArn, error) {
	svc := sts.New(session.New())

	// Get the MFA ARN
	input := &sts.GetCallerIdentityInput{}

	result, err := svc.GetCallerIdentity(input)
	if err != nil {
		return nil, err
	}

	callerIdentity := *result.Arn
	callerMFASerial := strings.Replace(callerIdentity, "user/", "mfa/", 1)

	token := &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int64(3600),
		SerialNumber:    aws.String(callerMFASerial),
		TokenCode:       aws.String(getMFAToken()),
	}

	resp, err := svc.GetSessionToken(token)

	if err != nil {
		return nil, err
	}

	awsCreds := &AwsCredentials{AccessKeyId: *resp.Credentials.AccessKeyId,
		SecretAccessKey: *resp.Credentials.SecretAccessKey,
		SessionToken:    *resp.Credentials.SessionToken,
		Expiration:      *resp.Credentials.Expiration}

	return &AwsCredentialsWithCallerArn{Credentials: awsCreds, IamArn: callerIdentity}, nil

}

func getIAMUserViaGetCallerIdentity() (string, error) {
	svc := sts.New(session.New())
	input := &sts.GetCallerIdentityInput{}

	result, err := svc.GetCallerIdentity(input)

	if err != nil {
		return "", err
	}

	pieces := strings.Split(*result.Arn, "/")
	return pieces[len(pieces)-1], nil
}

func getMFAToken() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter MFA code: ")
	text, _ := reader.ReadString('\n')

	return strings.TrimSpace(text)
}
