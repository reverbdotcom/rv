package main

import (
	"github.com/urfave/cli"

	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/sts"
	api "github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
)

type VaultToken struct {
	IamArn     string    `json:"iamArn"`
	Token      string    `json:"token"`
	Expiration time.Time `json:"expiration"`
}

type RDSEntry struct {
	Creds     string `json:"creds"`
	VaultRole string `json:"vaultRole"`
	IamRole   string `json:"iamRole"`
	DbName    string `json:"dbName"`
	Host      string `json:"host"`
}

type AwsCredentialsMap map[string]AwsCredentials

type VaultTokenMap map[string]VaultToken

type RDSEntryMap map[string]RDSEntry

type Decodeable interface {
	CacheLocation() string
}

func (t VaultToken) EncodeToString() string {
	b, _ := json.Marshal(t)
	return base64.StdEncoding.EncodeToString(b)
}

func (c AwsCredentials) EncodeToString() string {
	b, _ := json.Marshal(c)
	return base64.StdEncoding.EncodeToString(b)
}

func (c AwsCredentials) DecodeFromString(s string) AwsCredentials {
	var token AwsCredentials
	decoded, _ := base64.StdEncoding.DecodeString(s)
	json.Unmarshal(decoded, &token)
	return token
}

func (t VaultToken) DecodeFromString(s string) VaultToken {
	var token VaultToken
	decoded, _ := base64.StdEncoding.DecodeString(s)
	json.Unmarshal(decoded, &token)
	return token
}

func (VaultTokenMap) CacheLocation() string {
	return "cubbyhole/vault-tokens"
}

func (RDSEntryMap) CacheLocation() string {
	return "secret/rv/rds"
}

func (t VaultToken) expired() bool {
	return t.Expiration.Before(time.Now())
}

var baseVaultToken *VaultToken

func tokenConfigFileName() string {
	home := os.Getenv("HOME")
	return home + "/.rv/.token"
}

func createNewBaseToken() *VaultToken {
	c, _ := APIClient()
	creds, err := getSessionCredentials()
	if err != nil {
		return nil
	}
	token, err := getVaultToken(c, creds, "aws-user")
	if err != nil {
		return nil
	}
	baseVaultToken = token

	saveCachedVaultToken("aws-user", token)
	saveCachedAwsCredentials(creds.IamArn, creds.Credentials)
	return token
}

func loadBaseVaultToken() *VaultToken {
	// If we already have the token cached, get it
	if baseVaultToken != nil {
		return baseVaultToken
	}

	// Try and load it from disk
	var t *VaultToken = new(VaultToken)
	raw, err := ioutil.ReadFile(tokenConfigFileName())
	if err != nil {
		return createNewBaseToken()
	}

	err = json.Unmarshal(raw, t)

	if err != nil {
		return createNewBaseToken()
	}

	// Establish a baseline token and cache it
	if t == nil || (t.expired()) {
		return createNewBaseToken()
	}

	baseVaultToken = t
	return t
}

func RegisterVaultCommands(app *cli.App) {

	app.Commands = append(app.Commands, cli.Command{
		Name:     "vault",
		Usage:    "Vault",
		Category: "Vault",
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

func vaultServerAddress() string {
	return getConfigValueString("vault.address")
}

func vaultProxyAddress() string {
	return getConfigValueString("vault.proxy")
}

func readCachedAwsCredentialsList() AwsCredentialsMap {
	entryMap := make(AwsCredentialsMap)

	// This returns a map[string]interface{}
	result := readFromVault(entryMap)

	if result != nil {
		for k, v := range result {
			var token AwsCredentials
			entryMap[k] = token.DecodeFromString(v.(string))
		}
	}
	return entryMap
}

func readFromVault(d Decodeable) map[string]interface{} {
	c, err := APIClient()
	if err != nil {
		return nil
	}
	t := loadBaseVaultToken()
	if t == nil {
		return nil
	}
	c.SetToken(t.Token)
	result, err := c.Logical().Read(d.CacheLocation())
	if result == nil || err != nil {
		return nil
	}
	return result.Data
}

func readCachedVaultTokenList() VaultTokenMap {
	tokenMap := make(VaultTokenMap)

	// This returns a map[string]interface{}
	result := readFromVault(tokenMap)

	if result != nil {
		for k, v := range result {
			var token VaultToken
			tokenMap[k] = token.DecodeFromString(v.(string))
		}
	}
	return tokenMap
}

func readCachedVaultRDSList() RDSEntryMap {
	entryMap := make(RDSEntryMap)

	// This returns a map[string]interface{}
	result := readFromVault(entryMap)

	if result != nil {
		for k, v := range result {
			var entry RDSEntry
			mapstructure.Decode(v, &entry)
			entryMap[k] = entry
		}
	}
	return entryMap
}

func getCachedAwsCredentials(arn string) *AwsCredentials {
	creds := readCachedAwsCredentialsList()
	if cred, ok := creds[arn]; ok {
		if !cred.expired() {
			return &cred
		}
	}
	return nil
}

func saveCachedAwsCredentials(arn string, creds *AwsCredentials) {
	credList := readCachedAwsCredentialsList()
	credList[arn] = *creds
	//writeCachedAwsCredentialsList(credList)
	credList.SaveToVault()
}

func (m AwsCredentialsMap) CacheLocation() string {
	return "cubbyhole/aws-creds"
}

func saveToVault(d Decodeable, interfaceTokenMap map[string]interface{}) {
	c, _ := APIClient()
	t := loadBaseVaultToken()
	c.SetToken(t.Token)
	c.Logical().Write(d.CacheLocation(), interfaceTokenMap)
}

func (m AwsCredentialsMap) SaveToVault() {
	interfaceTokenMap := make(map[string]interface{})
	for k, v := range m {
		interfaceTokenMap[k] = v.EncodeToString()
	}

	saveToVault(m, interfaceTokenMap)
}

func (m VaultTokenMap) SaveToVault() {
	interfaceTokenMap := make(map[string]interface{})
	for k, v := range m {
		interfaceTokenMap[k] = v.EncodeToString()
	}

	saveToVault(m, interfaceTokenMap)
}

func getCachedVaultToken(vaultRole string) *VaultToken {
	t := loadBaseVaultToken()

	if vaultRole == "aws-user" {
		return t
	}
	tokenMap := readCachedVaultTokenList()
	for k, v := range tokenMap {
		if k == vaultRole {
			return &v
		}
	}

	return nil
}

func saveCachedVaultToken(vaultRole string, token *VaultToken) {
	if token == baseVaultToken {
		// We are saving the base vault token to disk
		jsonOut, _ := json.Marshal(token)
		ioutil.WriteFile(tokenConfigFileName(), jsonOut, 0600)
		return
	}

	tokenList := readCachedVaultTokenList()
	tokenList[vaultRole] = *token
	tokenList.SaveToVault()
}

//func auth(c *api.Client, creds *credentials.Credentials) error {
func getVaultToken(c *api.Client, creds *AwsCredentialsWithCallerArn, vaultRole string) (*VaultToken, error) {
	if c == nil {
		return nil, fmt.Errorf("vault api client is nil")
	}

	if creds == nil {
		return nil, fmt.Errorf("Error with AWS credentials")
	}

	svc, err := NewSTSService(creds.Credentials)

	if err != nil {
		return nil, err
	}

	var params *sts.GetCallerIdentityInput

	stsRequest, _ := svc.GetCallerIdentityRequest(params)
	stsRequest.Sign()

	headersJSON, err := json.Marshal(stsRequest.HTTPRequest.Header)
	if err != nil {
		return nil, err
	}
	requestBody, err := ioutil.ReadAll(stsRequest.HTTPRequest.Body)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	if secret == nil {
		return nil, fmt.Errorf("[vault] - empty response from credential provider")
	}

	tk := VaultToken{IamArn: creds.IamArn, Token: secret.Auth.ClientToken,
		Expiration: time.Now().Add(time.Duration(secret.Auth.LeaseDuration) * time.Second)}

	return &tk, nil
}

func APIClient() (*api.Client, error) {
	var httpClient *http.Client

	vaultAddress := vaultServerAddress()
	vaultProxy := vaultProxyAddress()

	if vaultProxy != "" {
		proxyURL, err := url.Parse(vaultProxy)
		if err != nil {
			return nil, err
		}

		httpClient = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}
	}

	config := &api.Config{
		Address:    vaultAddress,
		HttpClient: httpClient,
	}

	return api.NewClient(config)
}
