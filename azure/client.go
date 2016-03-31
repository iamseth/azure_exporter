package azure

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/prometheus/common/log"
)

// Client represents an API client to the Azure Resource Manager
// It handles wrapping the requests with the required authentication headers
type Client struct {
	Client         *http.Client
	Token          token
	SubscriptionID string
	Host           string
	Credentials    Credentials
}

// NewClient creates a new Azure API client based on a set of credentials represented by the azure.Credentials data type
func NewClient(c Credentials) (*Client, error) {
	endpoint := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", c.TenantID)
	tk, err := getToken(endpoint, c.ClientID, c.ClientSecret)
	if err != nil {
		log.Errorf("Failed to create a new Client while getting token: %s", err)
		return nil, err
	}

	return &Client{
		Client:         &http.Client{},
		Token:          tk,
		SubscriptionID: c.SubscriptionID,
		Host:           "management.azure.com",
		Credentials:    c,
	}, nil
}

// Credentials represents a set of Azure credentials for a service principal
// All fields are required
type Credentials struct {
	SubscriptionID string
	ClientID       string
	ClientSecret   string
	TenantID       string
}

// NewCredentialsFromFile will create a new instance of an azure.Credentials object from a JSON formatted file
func NewCredentialsFromFile(path string) Credentials {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Unable to create credentials from file %s: %s", path, err)
	}
	var creds Credentials
	json.Unmarshal(file, &creds)
	return creds
}

func (c Credentials) String() string {
	b, err := json.Marshal(c)
	if err != nil {
		log.Errorf("Cannot marshal Credentials: %s", err)
	}
	return string(b)
}

type token struct {
	Type      string      `json:"token_type"`
	Scope     string      `json:"scope"`
	Resource  string      `json:"resource"`
	Bearer    string      `json:"access_token"`
	ExpiresIn json.Number `json:"expires_in"`
	ExpiresOn json.Number `json:"expires_on"`
	NotBefore json.Number `json:"not_before"`
}

func getToken(endpoint, clientID, clientSecret string) (token, error) {
	log.Debugf("Getting authentication token")
	v := url.Values{}
	v.Set("grant_type", "client_credentials")
	v.Set("resource", "https://management.core.windows.net/")
	v.Set("client_id", clientID)
	v.Set("client_secret", clientSecret)

	var tk token
	resp, err := http.PostForm(endpoint, v)
	if err != nil {
		return tk, fmt.Errorf("Unable to get token: %s", err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&tk)
	if err != nil {
		return tk, fmt.Errorf("Unable to get token: %s", err)
	}

	log.Debugf("Recieved new ARM token")
	return tk, nil
}

func (c *Client) request(method, url string, body io.Reader) (*http.Response, error) {
	// check if auth is expired. if so, re-authenticate
	now := time.Now().UTC().Unix()
	expires, _ := c.Token.ExpiresOn.Int64()
	log.Debugf("It is now %d and the token will expire %d", now, expires)

	if expires < now {
		log.Debugf("Auth token has expired. Re-authenticating.")
		endpoint := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", c.Credentials.TenantID)
		token, err := getToken(endpoint, c.Credentials.ClientID, c.Credentials.ClientSecret)
		if err != nil {
			log.Errorf("Unable to authenticate with Azure: %s", err)
			return nil, err
		}
		c.Token = token
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token.Bearer))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) get(url string) (*http.Response, error) {
	return c.request("GET", url, nil)
}

func (c *Client) post(url string, body io.Reader) (*http.Response, error) {
	return c.request("POST", url, body)
}

func (c *Client) delete(url string) (*http.Response, error) {
	return c.request("DELETE", url, nil)
}

func (c *Client) put(url string, body io.Reader) (*http.Response, error) {
	return c.request("PUT", url, body)
}

func (c *Client) buildURL(group, resource, apiVersion string) url.URL {
	u := url.URL{Scheme: "https", Host: c.Host}
	if group != "" {
		group = "/resourceGroups/" + group

	}

	u.Path = fmt.Sprintf("/subscriptions/%s%s/providers/%s", c.SubscriptionID, group, resource)
	q := u.Query()
	q.Set("api-version", apiVersion)
	u.RawQuery = q.Encode()
	return u
}
