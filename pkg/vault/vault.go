package vault

import (
	"github.com/urfave/cli"

	"fmt"
	"io/ioutil"
	"encoding/base64"
	"encoding/json"
	"os"
	"net/url"
	"net/http"
	"time"

//	"github.com/aws/aws-sdk-go/aws"
//	"github.com/aws/aws-sdk-go/aws/credentials"
//	"github.com/aws/aws-sdk-go/aws/credentials"
//	"github.com/aws/aws-sdk-go/aws/session"
	api "github.com/hashicorp/vault/api"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/reverbdotcom/rv/pkg/iam"
)

type RoleTokens map[string]Token

type Token struct {
	Token string `json:"token"`
	Expiration time.Time `json:"expiration"`
}

func configFileName() string {
	home := os.Getenv("HOME")
	return home + "/.rv.json"
	
}

func getConfig() RoleTokens {
	var rt RoleTokens
	raw, err := ioutil.ReadFile(configFileName())
	if err != nil {
		return rt
	}

	json.Unmarshal(raw, &rt)
	return rt
}

func saveToken(role string, token Token) {
	rt := getConfig()
	if rt == nil {
		rt = make(RoleTokens)
	}
	rt[role] = token
	jsonOut, _ := json.Marshal(rt)
	ioutil.WriteFile(configFileName(), jsonOut, 0400)
}

func RegisterCommands(app *cli.App) {

	app.Commands = append(app.Commands, cli.Command{
		Name:		"vault",
		Usage:		"Vault",
		Category:	"Vault",
		Subcommands: []cli.Command{
			{
				Name: "auth",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
		},
	})
}

func GetCachedVaultToken(vaultRole string) (string) {
	rt := getConfig()

	if val, ok := rt[vaultRole]; ok {
		token := val.Token
		expiry := val.Expiration

		// Token hasn't yet expired, return saved token
		if(time.Now().Before(expiry)) {
			return token
		}
	}

	return ""
}

//func auth(c *api.Client, creds *credentials.Credentials) error {
func GetVaultToken(c *api.Client, creds *sts.Credentials, vaultRole string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("api client is nil")
	}

	svc, err := iam.NewSTSService(creds)

	if err != nil {
		return "", err
	}

	var params *sts.GetCallerIdentityInput

	stsRequest, _ := svc.GetCallerIdentityRequest(params)
	stsRequest.Sign()

	headersJSON, err := json.Marshal(stsRequest.HTTPRequest.Header)
	if err != nil {
		return "", err
	}
	requestBody, err := ioutil.ReadAll(stsRequest.HTTPRequest.Body)
	if err != nil {
		return "", err
	}

	method := stsRequest.HTTPRequest.Method
	targetURL := base64.StdEncoding.EncodeToString([]byte(stsRequest.HTTPRequest.URL.String()))
	headers := base64.StdEncoding.EncodeToString(headersJSON)
	body := base64.StdEncoding.EncodeToString(requestBody)

	// And pass them on to the Vault server
	secret, err := c.Logical().Write("auth/aws/login", map[string]interface{}{
		"iam_http_request_method": method,
		"iam_request_url":         targetURL,
		"iam_request_headers":     headers,
		"iam_request_body":        body,
        "role":                    vaultRole,
	})

	if err != nil {
		return "", err
	}

	if secret == nil {
		return "", fmt.Errorf("[vault] - empty response from credential provider")
	}
	
	saveToken(vaultRole, Token{Token: secret.Auth.ClientToken, 
		Expiration: time.Now().Add(time.Duration(secret.Auth.LeaseDuration)*time.Second)})

	return secret.Auth.ClientToken, nil
}

func ApiClient() (*api.Client, error) {
	var httpClient *http.Client

	vaultAddress := "https://vault.reverb.com"
	overrideAddr, overrideExists := os.LookupEnv("VAULT_ADDR")

	if overrideExists {
		vaultAddress = overrideAddr
	}

	vaultProxy := os.Getenv("VAULT_PROXY")

	if vaultProxy != "" {
		proxyUrl, err := url.Parse(vaultProxy)
		if err != nil {
			return nil, err
		}

		httpClient = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
			},
		}
	}

	config := &api.Config{
		Address: vaultAddress,
		HttpClient: httpClient,
	}

	return api.NewClient(config)
}
