package card

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/carddav"
	"github.com/rstms/mabctl/util"
	"net/http"
	"regexp"
	"strings"
)

type Client struct {
	username string
	password string
	url      string
	client   *carddav.Client
}

type digestAuthHTTPClient struct {
	c           webdav.HTTPClient
	username    string
	password    string
	nonce       string
	opaque      string
	realm       string
	qop         string
	nonceCount  int
	clientNonce string
	hash1       string
}

func (c *digestAuthHTTPClient) Do(req *http.Request) (*http.Response, error) {
	fmt.Printf("\nrequest: %+v\n", req)
	if c.realm != "" {
		c.generateAuthHeader(req)
	}
	resp, err := c.c.Do(req)
	if err != nil {
		return nil, util.Fatalf("request failed: %v\n", err)
	}
	if resp.StatusCode == 401 {
		authHeader, ok := resp.Header["Www-Authenticate"]
		if !ok {
			return nil, util.Fatalf("Received %s w/o WWW-Authenticate header", resp.Status)
		}
		err := c.handleAuthHeader(authHeader)
		if err != nil {
			return nil, err
		}
		if c.realm != "" {
			c.generateAuthHeader(req)
			resp, err := c.c.Do(req)
			if err != nil {
				return nil, util.Fatalf("authorized request failed: %v", err)
			}
			return resp, nil
		}
		panic("realm not set")
	}
	fmt.Printf("\nresponse: %+v\n", resp)
	return resp, err
}

// var AUTH_HEADER_PATTERN = regexp.MustCompile(`^(Digest) \([^=]*="[^"]*\)",*.*`)
// var AUTH_HEADER_PATTERN = regexp.MustCompile(`([a-zA-Z][^=]*="[^"]")`)
var HEADER = regexp.MustCompile(`([a-zA-Z][^=]*="[^"]*")`)
var FIELD = regexp.MustCompile(`([^=]*)="([^"]*)"`)

func (c *digestAuthHTTPClient) handleAuthHeader(values []string) error {
	if len(values) != 1 {
		return util.Fatalf("unexpected auth header format")
	}
	if values[0][0:7] != "Digest " {
		return util.Fatalf("auth header not Digest")
	}
	matches := HEADER.FindAllStringSubmatch(values[0][7:], -1)
	for _, match := range matches {
		fields := FIELD.FindStringSubmatch(match[1])
		if len(fields) != 3 {
			return util.Fatalf("failed parsing auth header fields")
		}
		switch fields[1] {
		case "nonce":
			c.nonce = fields[2]
			c.nonceCount = 0
			n, err := nonce(16)
			if err != nil {
				return util.Fatalf("failed generating nonce: %v", err)
			}
			c.clientNonce = n
		case "opaque":
			c.opaque = fields[2]
		case "qop":
			c.qop = fields[2]
		case "realm":
			c.realm = fields[2]
			c.hash1 = hash(c.username + ":" + c.realm + ":" + c.password)
		default:
			return util.Fatalf("unexpected field in auth header: %s", fields[1])
		}
	}
	return nil
}

func nonce(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	n := base64.StdEncoding.EncodeToString(randomBytes)
	return n, nil
}

func hash(clearText string) string {
	hasher := md5.New()
	hasher.Write([]byte(clearText))
	hash := hasher.Sum(nil)
	hashText := hex.EncodeToString(hash)
	fmt.Printf("hash(%s)->%s\n", clearText, hashText)
	return hashText
}

func (c *digestAuthHTTPClient) generateAuthHeader(req *http.Request) {
	c.nonceCount++
	hash2 := hash(req.Method + ":" + req.URL.Path)
	response := hash(c.hash1 + ":" + c.nonce + ":" + hash2)
	header := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s" response="%s", qop=%s, nc=%d, cnonce="%s"`,
		c.username, c.realm, c.nonce, req.RequestURI, response, c.qop, c.nonceCount, c.clientNonce)
	fmt.Printf("Setting Authorization header: %s\n", header)
	req.Header.Set("Authorization", header)
}

func HTTPClientWithDigestAuth(client *http.Client, username, password string) webdav.HTTPClient {
	d := digestAuthHTTPClient{c: client, username: username, password: password}
	if client == nil {
		d.c = http.DefaultClient
	}
	return &d
}

func NewClient(username, password, url string) (*Client, error) {
	fmt.Printf("%s %s %s\n", username, password, url)
	if url == "" {
		fields := strings.Split(username, "@")
		if len(fields) != 2 {
			return nil, util.Fatalf("invalid email address format: %s", username)
		}
		domain := fields[1]
		u, err := carddav.DiscoverContextURL(context.Background(), domain)
		if err != nil {
			return nil, util.Fatalf("failed carddav URL discovery for domain %s :%v", domain, err)
		}
		url = u
		fmt.Printf("discovered url: %s\n", url)
	}

	c := Client{username, password, url, nil}

	//httpClient := webdav.HTTPClientWithBasicAuth(nil, username, password)
	httpClient := HTTPClientWithDigestAuth(nil, username, password)
	cdc, err := carddav.NewClient(httpClient, url)
	if err != nil {
		return nil, util.Fatalf("failed creating webdav client: %v", err)
	}
	c.client = cdc

	return &c, nil
}

func (c *Client) List() (string, error) {
	homeSet, err := c.client.FindAddressBookHomeSet(context.Background(), c.username)
	if err != nil {
		return "", util.Fatalf("FindAddressBookHomeSet failed: %v", err)
	}
	return homeSet, nil
}

/*
func (c *Client) readBody(method, url string, resp *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, util.Fatalf("%s %s failed reading response body: %v", method, url, err)
	}

	return body, nil
}

func (c *Client) get(url string) ([]byte, error) {
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, util.Fatalf("GET %s failed: %v", url, err)
	}
	defer resp.Body.Close()
	return c.readBody("GET", url, resp)
}

func (c *Client) post(url string, data *bytes.Buffer) ([]byte, error) {
	resp, err := c.client.Post(url, "application/json", data)
	if err != nil {
		return nil, util.Fatalf("POST %s failed: %v", url, err)
	}
	defer resp.Body.Close()
	return c.readBody("POST", url, resp)
}

func (c *Client) del(url string) ([]byte, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, util.Fatalf("failed creating DELETE %s request: %v", url, err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, util.Fatalf("DELETE %s failed: %v", url, err)
	}
	defer resp.Body.Close()
	return c.readBody("DELETE", url, resp)
}

func (c *Client) formatResponse(body []byte) (string, error) {
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return "", util.Fatalf("failed decoding response: %v", err)
	}
	formatted, err := json.MarshalIndent(&data, "", "  ")
	if err != nil {
		return "", util.Fatalf("failed formatting response: %v", err)
	}
	return string(formatted), nil
}

func (c *Client) Initialize() (string, error) {

	body, err := c.post(c.url+"/initialize/", bytes.NewBuffer([]byte{}))
	if err != nil {
		return "", err
	}
	return c.formatResponse(body)
}

func (c *Client) Reset() (string, error) {

	body, err := c.post(c.url+"/reset/", bytes.NewBuffer([]byte{}))
	if err != nil {
		return "", err
	}
	return c.formatResponse(body)
}

func (c *Client) GetStatus() (string, error) {

	body, err := c.get(c.url + "/status/")
	if err != nil {
		return "", err
	}
	return c.formatResponse(body)
}

func (c *Client) GetUsers() (string, error) {
	body, err := c.get(c.url + "/users/")
	if err != nil {
		return "", err
	}
	return c.formatResponse(body)
}

func (c *Client) GetAddressBooks(email string) (string, error) {
	path := fmt.Sprintf("%s/addressbooks/%s/", c.url, url.PathEscape(email))
	body, err := c.get(path)
	if err != nil {
		return "", err
	}
	return c.formatResponse(body)
}

func (c *Client) AddUser(email, display, password string) (string, error) {
	user := map[string]string{
		"username":    email,
		"displayname": display,
		"password":    password,
	}
	jsonData, err := json.Marshal(user)
	if err != nil {
		return "", util.Fatalf("failed formatting form data: %v", err)
	}
	body, err := c.post(c.url+"/user/", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", util.Fatalf("POST failed: %v", err)
	}
	return c.formatResponse(body)
}

func (c *Client) AddAddressBook(email, name, description string) (string, error) {
	user := map[string]string{
		"username":    email,
		"bookname":    name,
		"description": description,
	}
	jsonData, err := json.Marshal(user)
	if err != nil {
		return "", util.Fatalf("failed formatting form data: %v", err)
	}
	body, err := c.post(c.url+"/addressbook/", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", util.Fatalf("POST failed: %v", err)
	}
	return c.formatResponse(body)
}

func (c *Client) DeleteUser(email string) (string, error) {
	path := fmt.Sprintf("%s/user/%s/", c.url, url.PathEscape(email))
	body, err := c.del(path)
	if err != nil {
		return "", util.Fatalf("DELETE failed: %v", err)
	}
	return c.formatResponse(body)
}

func (c *Client) DeleteAddressBook(userEmail, addressBookName string) (string, error) {
	path := fmt.Sprintf("%s/addressbook/%s/%s/", c.url, url.PathEscape(userEmail), url.PathEscape(addressBookName))
	body, err := c.del(path)
	if err != nil {
		return "", util.Fatalf("DELETE failed: %v", err)
	}
	return c.formatResponse(body)
}
*/
