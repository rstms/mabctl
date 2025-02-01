package admin

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rstms/mabctl/util"
	"io/ioutil"
	"net/http"
	"net/url"
)

type ErrorMessage struct {
	Error string `json:"error"`
}

type Client struct {
	username string
	password string
	url      string
	cert     string
	key      string
	apikey   string
	client   *http.Client
}

type User struct {
	UserName    string `json:"username"`
	DisplayName string `json:"displayname"`
	URI         string `json:"uri"`
}

type AddressBook struct {
	UserName    string `json:"username"`
	BookName    string `json:"bookname"`
	Description string `json:"description"`
	Contacts    int    `json:"contacts"`
	Token       string `json:"token"`
	URI         string `json:"uri"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Request string `json:"request"`
}

type UsersResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Request string `json:"request"`
	Users   []User `json:"users"`
}

func NewClient(username, password, url, cert, key, apikey string, insecure bool) (*Client, error) {
	c := Client{username, password, url, cert, key, apikey, nil}

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

func (c *Client) request(method, path string, data *[]byte) (*http.Request, error) {
	fmt.Printf("request: %s, %s, %v\n", method, path, data)
	var body *bytes.Buffer
	if data == nil {
		body = bytes.NewBuffer([]byte{})
	} else {
		body = bytes.NewBuffer(*data)
	}
	req, err := http.NewRequest(method, c.url+path, body)
	if err != nil {
		return nil, util.Fatalf("failed creating %s request: %v", method, err)
	}
	req.Header.Set("X-Api-Key", c.apikey)
	req.Header.Set("X-Admin-Username", c.username)
	req.Header.Set("X-Admin-Password", c.password)
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (c *Client) get(path string, ret *map[string]interface{}) (string, *map[string]interface{}, error) {
	req, err := c.request("GET", path, nil)
	if err != nil {
		return "", nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return "", nil, util.Fatalf("GET %s failed: %v", path, err)
	}
	defer resp.Body.Close()
	return c.handleResponse("GET", path, resp, ret)
}

func (c *Client) post(path string, data *[]byte, ret *map[string]interface{}) (string, *map[string]interface{}, error) {
	req, err := c.request("POST", path, data)
	if err != nil {
		return "", nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return "", nil, util.Fatalf("POST %s failed: %v", path, err)
	}
	defer resp.Body.Close()
	return c.handleResponse("POST", path, resp, ret)
}

func (c *Client) del(path string, data *[]byte, ret *map[string]interface{}) (string, *map[string]interface{}, error) {
	req, err := c.request("DELETE", path, data)
	if err != nil {
		return "", nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return "", nil, util.Fatalf("DELETE %s failed: %v", path, err)
	}
	defer resp.Body.Close()
	return c.handleResponse("DELETE", path, resp, ret)
}

func (c *Client) handleResponse(method, path string, resp *http.Response, ret *map[string]interface{}) (string, *map[string]interface{}, error) {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", nil, util.Fatalf("Error: %s %s '%s'", method, path, resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, util.Fatalf("%s %s failed reading response body: %v", method, path, err)
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
	return c.post("/initialize/", nil, &ret)
}

func (c *Client) Reset() (string, *map[string]interface{}, error) {
	var ret map[string]interface{}
	return c.post("/reset/", nil, &ret)
}

func (c *Client) GetStatus() (string, *map[string]interface{}, error) {
	var ret map[string]interface{}
	return c.get("/status/", &ret)
}

func (c *Client) GetUptime() (string, *map[string]interface{}, error) {
	var ret map[string]interface{}
	return c.get("/uptime/", &ret)
}

func (c *Client) RequestShutdown() (string, *map[string]interface{}, error) {
	var ret map[string]interface{}
	return c.post("/shutdown/", nil, &ret)
}

func (c *Client) checkError(result string) error {
	var msg ErrorMessage
	err := json.Unmarshal([]byte(result), &msg)
	if err != nil {
		return nil
	}
	return errors.New(msg.Error)
}

func (c *Client) GetUsers() (string, *[]User, error) {
	var ret map[string]interface{}
	formatted, result, err := c.get("/users/", &ret)
	if err != nil {
		return "", nil, err
	}

	fmt.Println(result)

	//fmt.Printf("BEGIN RESULT\n%vEND RESULT\n", result)
	/*
		err = c.checkError(formatted)
			if err != nil {
				return "", nil, util.Fatalf("%v", err)
			}

			parsed := UsersResponse{}
			err = json.Unmarshal([]byte(formatted), &parsed)
			if err != nil {
				return "", nil, util.Fatalf("failed unmarshalling users JSON response: %v", err)
			}
	*/
	var users []User
	return formatted, &users, nil
}

func (c *Client) GetAddressBooks(username string) (string, *map[string]AddressBook, error) {
	var ret map[string]interface{}
	var formatted string
	var err error
	if username == "" {
		formatted, _, err = c.get("/books/", &ret)
	} else {
		formatted, _, err = c.get(fmt.Sprintf("/books/%s/", url.PathEscape(username)), &ret)
	}
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

func (c *Client) AddUser(username, display, password string) (string, *map[string]interface{}, error) {
	user := map[string]string{
		"username":    username,
		"displayname": display,
		"password":    password,
	}
	jsonData, err := json.Marshal(user)
	if err != nil {
		return "", nil, util.Fatalf("failed formatting add user request data: %v", err)
	}
	var ret map[string]interface{}
	return c.post("/user/", &jsonData, &ret)
}

func (c *Client) AddAddressBook(email, name, description string) (string, *map[string]interface{}, error) {
	book := map[string]string{
		"username":    email,
		"bookname":    name,
		"description": description,
	}
	jsonData, err := json.Marshal(book)
	if err != nil {
		return "", nil, util.Fatalf("failed formatting add book request data: %v", err)
	}
	var ret map[string]interface{}
	return c.post("/book/", &jsonData, &ret)
}

func (c *Client) DeleteUser(email string) (string, *map[string]interface{}, error) {
	user := map[string]string{
		"username": email,
	}
	jsonData, err := json.Marshal(user)
	if err != nil {
		return "", nil, util.Fatalf("failed formatting delete user request data: %v", err)
	}
	var ret map[string]interface{}
	return c.del("/user/", &jsonData, &ret)
}

func (c *Client) DeleteAddressBook(username, token string) (string, *map[string]interface{}, error) {
	user := map[string]string{
		"username": username,
		"token":    token,
	}
	jsonData, err := json.Marshal(user)
	if err != nil {
		return "", nil, util.Fatalf("failed formatting delete user request data: %v", err)
	}

	var ret map[string]interface{}
	return c.del("/book/", &jsonData, &ret)
}
