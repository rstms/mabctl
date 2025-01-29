package admin

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/rstms/mabctl/util"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Client struct {
	username string
	password string
	url      string
	cert     string
	key      string
	client   *http.Client
}

type User struct {
	DisplayName  string `json:"display_name"`
	EmailAddress string `json:"email"`
}

type AddressBook struct {
	Contacts    int    `json:"contacts"`
	Description string `json:"description"`
}

func NewClient(username, password, url, cert, key string, insecure bool) (*Client, error) {
	c := Client{username, password, url, cert, key, nil}

	clientCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, util.Fatalf("failed loading client certificate: %v", err)
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{clientCert},
		InsecureSkipVerify: insecure,
	}
	c.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return &c, nil
}

func (c *Client) get(url string, ret *map[string]interface{}) (string, *map[string]interface{}, error) {
	resp, err := c.client.Get(url)
	if err != nil {
		return "", nil, util.Fatalf("GET %s failed: %v", url, err)
	}
	defer resp.Body.Close()
	return c.handleResponse("GET", url, resp, ret)
}

func (c *Client) post(url string, data *bytes.Buffer, ret *map[string]interface{}) (string, *map[string]interface{}, error) {
	resp, err := c.client.Post(url, "application/json", data)
	if err != nil {
		return "", nil, util.Fatalf("POST %s failed: %v", url, err)
	}
	defer resp.Body.Close()
	return c.handleResponse("POST", url, resp, ret)
}

func (c *Client) del(url string, ret *map[string]interface{}) (string, *map[string]interface{}, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return "", nil, util.Fatalf("failed creating DELETE %s request: %v", url, err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return "", nil, util.Fatalf("DELETE %s failed: %v", url, err)
	}
	defer resp.Body.Close()
	return c.handleResponse("DELETE", url, resp, ret)
}

func (c *Client) handleResponse(method, url string, resp *http.Response, ret *map[string]interface{}) (string, *map[string]interface{}, error) {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", nil, util.Fatalf("Error: %s %s '%s'", method, url, resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, util.Fatalf("%s %s failed reading response body: %v", method, url, err)
	}
	return c.formatResponse(body, ret)
}

func (c *Client) formatResponse(body []byte, ret *map[string]interface{}) (string, *map[string]interface{}, error) {
	if len(body) == 0 {
		return "{}", nil, nil
	}
	err := json.Unmarshal(body, ret)
	if err != nil {
		fmt.Printf("body: '%s'\n", string(body))
		return "", nil, util.Fatalf("failed decoding response: %v", err)
	}
	formatted, err := c.Format(ret)
	if err != nil {
		return "", nil, err
	}
	return formatted, ret, nil
}

func (c *Client) Format(data interface{}) (string, error) {
	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", util.Fatalf("failed formatting JSON: %v", err)
	}
	return string(formatted), nil
}

func (c *Client) Initialize() (string, *map[string]interface{}, error) {
	var ret map[string]interface{}
	return c.post(c.url+"/initialize/", bytes.NewBuffer([]byte{}), &ret)
}

func (c *Client) Reset() (string, *map[string]interface{}, error) {
	var ret map[string]interface{}
	return c.post(c.url+"/reset/", bytes.NewBuffer([]byte{}), &ret)
}

func (c *Client) GetStatus() (string, *map[string]interface{}, error) {
	var ret map[string]interface{}
	return c.get(c.url+"/status/", &ret)
}

func (c *Client) GetUsers() (string, *map[string]User, error) {
	var ret map[string]interface{}
	formatted, _, err := c.get(c.url+"/users/", &ret)
	if err != nil {
		return "", nil, err
	}
	var users map[string]User
	err = json.Unmarshal([]byte(formatted), &users)
	if err != nil {
		return "", nil, util.Fatalf("failed encoding Users: %v", err)
	}
	return formatted, &users, nil
}

func (c *Client) GetAddressBooks(email string) (string, *map[string]AddressBook, error) {
	var ret map[string]interface{}
	formatted, _, err := c.get(fmt.Sprintf("%s/addressbooks/%s/", c.url, url.PathEscape(email)), &ret)
	if err != nil {
		return "", nil, err
	}
	var books map[string]AddressBook
	err = json.Unmarshal([]byte(formatted), &books)
	if err != nil {
		return "", nil, util.Fatalf("failed encoding AddressBooks: %v", err)
	}
	return formatted, &books, nil
}

func (c *Client) AddUser(email, display, password string) (string, *map[string]interface{}, error) {
	user := map[string]string{
		"username":    email,
		"displayname": display,
		"password":    password,
	}
	jsonData, err := json.Marshal(user)
	if err != nil {
		return "", nil, util.Fatalf("failed formatting form data: %v", err)
	}
	var ret map[string]interface{}
	return c.post(c.url+"/user/", bytes.NewBuffer(jsonData), &ret)
}

func (c *Client) AddAddressBook(email, name, description string) (string, *map[string]interface{}, error) {
	book := map[string]string{
		"username":    email,
		"bookname":    name,
		"description": description,
	}
	jsonData, err := json.Marshal(book)
	if err != nil {
		return "", nil, util.Fatalf("failed formatting form data: %v", err)
	}
	var ret map[string]interface{}
	return c.post(c.url+"/addressbook/", bytes.NewBuffer(jsonData), &ret)
}

func (c *Client) DeleteUser(email string) (string, *map[string]interface{}, error) {
	var ret map[string]interface{}
	return c.del(fmt.Sprintf("%s/user/%s/", c.url, url.PathEscape(email)), &ret)
}

func (c *Client) DeleteAddressBook(userEmail, addressBookName string) (string, *map[string]interface{}, error) {
	var ret map[string]interface{}
	return c.del(fmt.Sprintf("%s/addressbook/%s/%s/", c.url, url.PathEscape(userEmail), url.PathEscape(addressBookName)), &ret)
}
