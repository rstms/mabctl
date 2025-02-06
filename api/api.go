package api

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

type Book struct {
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
	Response
	Users []User `json:"users"`
}

type AddUserResponse struct {
	Response
	User User `json:"user"`
}

type BooksResponse struct {
	Response
	Books []Book `json:"books"`
}

type AddBookResponse struct {
	Response
	Book Book `json:"book"`
}

type StatusResponse struct {
	Response
	Status map[string]string `json:"status"`
}

type ErrorResponse struct {
	Response
	Detail string `json:"detail"`
}

func Format(data interface{}) (string, error) {
	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", util.Fatalf("failed formatting JSON: %v", err)
	}
	return string(formatted), nil
}

func FormatIfJSON(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	decoded := map[string]interface{}{}
	err := json.Unmarshal(body, &decoded)
	if err != nil {
		return string(body)
	}
	formatted, err := json.MarshalIndent(&decoded, "", "  ")
	if err != nil {
		return string(body)
	}
	return string(formatted)
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

func (c *Client) get(path string, ret interface{}) error {
	req, err := c.request("GET", path, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return util.Fatalf("GET %s failed: %v", path, err)
	}
	defer resp.Body.Close()
	return c.handleResponse("GET", path, resp, ret)
}

func (c *Client) post(path string, data *[]byte, ret interface{}) error {
	req, err := c.request("POST", path, data)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return util.Fatalf("POST %s failed: %v", path, err)
	}
	defer resp.Body.Close()
	return c.handleResponse("POST", path, resp, ret)
}

func (c *Client) del(path string, data *[]byte, ret interface{}) error {
	req, err := c.request("DELETE", path, data)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return util.Fatalf("DELETE %s failed: %v", path, err)
	}
	defer resp.Body.Close()
	return c.handleResponse("DELETE", path, resp, ret)
}

func (c *Client) handleResponse(method, path string, resp *http.Response, ret interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return util.Fatalf("%s %s failed reading response body: %v", method, path, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return util.Fatalf("Error: %s %s '%s'\n%s", method, path, resp.Status, FormatIfJSON(body))
	}
	if len(body) == 0 {
		return nil
	}
	err = json.Unmarshal(body, ret)
	if err != nil {
		return util.Fatalf("failed decoding response: %v\n%v", err, string(body))
	}
	return nil
}

func (c *Client) Initialize() (*Response, error) {
	var ret Response
	err := c.post("/initialize/", nil, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Client) Reset() (*Response, error) {
	var ret Response
	err := c.post("/reset/", nil, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Client) GetStatus() (*StatusResponse, error) {
	var ret StatusResponse
	err := c.get("/status/", &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Client) GetUptime() (*Response, error) {
	var ret Response
	err := c.get("/uptime/", &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Client) RequestShutdown() (*Response, error) {
	var ret Response
	err := c.post("/shutdown/", nil, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Client) GetUsers() (*UsersResponse, error) {
	var ret UsersResponse
	err := c.get("/users/", &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Client) GetBooks(username string) (*BooksResponse, error) {
	var ret BooksResponse
	var err error
	if username == "" {
		err = c.get("/books/", &ret)
	} else {
		err = c.get(fmt.Sprintf("/books/%s/", url.PathEscape(username)), &ret)
	}
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Client) AddUser(username, display, password string) (*AddUserResponse, error) {
	user := map[string]string{
		"username":    username,
		"displayname": display,
		"password":    password,
	}
	jsonData, err := json.Marshal(user)
	if err != nil {
		return nil, util.Fatalf("failed formatting add user request data: %v", err)
	}
	var ret AddUserResponse
	err = c.post("/user/", &jsonData, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Client) AddBook(email, name, description string) (*AddBookResponse, error) {
	book := map[string]string{
		"username":    email,
		"bookname":    name,
		"description": description,
	}
	jsonData, err := json.Marshal(book)
	if err != nil {
		return nil, util.Fatalf("failed formatting add book request data: %v", err)
	}
	var ret AddBookResponse
	err = c.post("/book/", &jsonData, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Client) DeleteUser(email string) (*Response, error) {
	user := map[string]string{
		"username": email,
	}
	jsonData, err := json.Marshal(user)
	if err != nil {
		return nil, util.Fatalf("failed formatting delete user request data: %v", err)
	}
	var ret Response
	err = c.del("/user/", &jsonData, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Client) DeleteBook(username, token string) (*Response, error) {
	user := map[string]string{
		"username": username,
		"token":    token,
	}
	jsonData, err := json.Marshal(user)
	if err != nil {
		return nil, util.Fatalf("failed formatting delete user request data: %v", err)
	}
	var ret Response
	err = c.del("/book/", &jsonData, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}
