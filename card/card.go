package card

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"github.com/rstms/go-webdav"
	"github.com/rstms/go-webdav/carddav"
	"github.com/rstms/mabctl/util"
	"net/http"
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
	method      string
	uri         string
}

func (c *digestAuthHTTPClient) Do(req *http.Request) (*http.Response, error) {
	fmt.Printf("\nrequest: %+v\n", req)
	fmt.Printf("\nheaders: %+v\n", req.Header)
	if c.realm != "" {
		panic("regenerate")
		c.generateAuthHeader(req)
	}
	resp, err := c.c.Do(req)
	if err != nil {
		return nil, util.Fatalf("request failed: %v\n", err)
	}
	if resp.StatusCode == http.StatusUnauthorized {

		challengeHeader := resp.Header.Get("Www-Authenticate")

		err := c.parseAuthHeader(challengeHeader)
		if err != nil {
			return nil, err
		}

		err = c.generateAuthHeader(req)
		if err != nil {
			return nil, err
		}
		resp, err = c.c.Do(req)
		if err != nil {
			fmt.Printf("\nresponse: %v\n", resp)
			return nil, util.Fatalf("response-authorized request failed: %v", err)
		}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, util.Fatalf("request failed: [%d] %v\n", resp.StatusCode, resp)
	}
	return resp, nil
}

func (c *digestAuthHTTPClient) parseAuthHeader(authHeader string) error {
	fmt.Printf("challengeHeader: %s\n", authHeader)

	if !strings.HasPrefix(authHeader, "Digest ") {
		return util.Fatalf("unrecognised auth header: %s", authHeader)
	}
	parts := strings.SplitAfter(authHeader, " ")
	authHeader = parts[1]
	c.nonce = ""
	c.realm = ""
	c.opaque = ""
	c.qop = ""
	c.uri = ""
	for _, param := range strings.Split(authHeader, ",") {
		parts := strings.SplitN(strings.TrimSpace(param), "=", 2)
		if len(parts) == 2 {
			key := strings.Trim(parts[0], ` "`)
			value := strings.Trim(parts[1], ` "`)
			fmt.Printf("%s: %s\n", key, value)
			switch key {
			case "nonce":
				c.nonce = value
			case "realm":
				c.realm = value
			case "opaque":
				c.opaque = value
			case "qop":
				c.qop = value
			case "uri":
				c.uri = value
			}
		}
	}
	if c.nonce == "" {
		return util.Fatalf("nonce not found: %v", authHeader)
	}
	if c.realm == "" {
		return util.Fatalf("realm not found: %v", authHeader)
	}
	if c.qop != "auth" {
		return util.Fatalf("unimplemented qop: %v", c.qop)
	}
	return nil
}

func randomNonce(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", util.Fatalf("random nonce generation failed: %v", err)
	}
	return fmt.Sprintf("%x", string(randomBytes)), nil
}

func md5sum(data string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

func encodeNonce(count int) string {
	return fmt.Sprintf("%08x", count)
}

func (c *digestAuthHTTPClient) generateAuthHeader(req *http.Request) error {

	c.method = req.Method
	//c.uri = req.URL.Path

	c.nonceCount++
	n, err := randomNonce(6)
	if err != nil {
		return err
	}
	c.clientNonce = n

	/* example values from RFC2617
	c.username="Mufasa"
	c.password="Circle Of Life"
	c.realm="testrealm@host.com"
	c.method="GET"
	c.uri="/dir/index.html"
	c.nonce = "dcd98b7102dd2f0e8b11d0f600bfb0c093"
	c.opaque ="5ccc069c403ebaf9f0171e9517f40e41"
	c.clientNonce = "0a4f113b"
	c.nonceCount=1
	*/

	HA1Text := c.username + ":" + c.realm + ":" + c.password
	HA1 := md5sum(HA1Text)
	fmt.Printf("HA1: %s %s\n", HA1, HA1Text)

	HA2Text := fmt.Sprintf("%s:%s", c.method, c.uri)
	HA2 := md5sum(HA2Text)
	fmt.Printf("HA2: %s %s\n", HA2, HA2Text)

	responseText := fmt.Sprintf("%s:%s:%s:%s:%s:%s", HA1, c.nonce, encodeNonce(c.nonceCount), c.clientNonce, c.qop, HA2)
	response := md5sum(responseText)
	fmt.Printf("response: %s %s\n", response, responseText)

	header := fmt.Sprintf(`Digest username="%s",realm="%s",nonce="%s",uri="%s",response="%s",qop="%s",nc="%s",cnonce="%s",opaque="%s"`,
		c.username, c.realm, c.nonce, c.uri, response, c.qop, encodeNonce(c.nonceCount), c.clientNonce, c.opaque)
	fmt.Printf("responseHeader: %s\n", header)
	req.Header.Set("Authorization", header)
	return nil
}

func HTTPClientWithDigestAuth(client *http.Client, username, password string) webdav.HTTPClient {
	d := digestAuthHTTPClient{c: client, username: username, password: password}
	if client == nil {
		d.c = http.DefaultClient
	}
	return &d
}

func NewClient(username, password, url string) (*Client, error) {
	fmt.Printf("NewClient: %s %s %s\n", username, password, url)
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
	ctx := context.Background()
	cup, err := c.client.FindCurrentUserPrincipal(ctx)
	if err != nil {
		return "", util.Fatalf("FindCurrentUserPrincipal failed: %v", err)
	}
	fmt.Print("CurrentUserPrincipal: %+v\n", cup)
	homeSet, err := c.client.FindAddressBookHomeSet(context.Background(), cup)
	if err != nil {
		return "", util.Fatalf("FindAddressBookHomeSet failed: %v", err)
	}
	return homeSet, nil
}
