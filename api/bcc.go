package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/emersion/go-webdav/carddav"
	"github.com/rstms/mabctl/util"
	"io/ioutil"
	"net/http"
	"net/url"
)

const PASSWORD_LENGTH = 12

type ErrorMessage struct {
	Error string `json:"error"`
}

type Controller struct {
	username string
	password string
	url      string
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

type AdminBooksResponse struct {
	Response
	Books []Book `json:"books"`
}

type BooksResponse struct {
	Response
	Books []carddav.AddressBook `json:"books"`
}

func (b *BooksResponse) Names() ([]string, error) {
	ret := make([]string, len(b.Books))
	for i, book := range b.Books {
		_, name, err := util.ParseBookPath("", book.Path)
		if err != nil {
			return []string{}, err
		}
		ret[i] = name
	}
	return ret, nil
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

type AddressesResponse struct {
	Response
	Addresses []carddav.AddressObject `json:"addresses"`
}

type AddressResponse struct {
	Response
	Address carddav.AddressObject `json:"address"`
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

func (c *Controller) request(method, path string, data *[]byte) (*http.Request, error) {
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

func (c *Controller) get(path string, ret interface{}) error {
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

func (c *Controller) post(path string, data *[]byte, ret interface{}) error {
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

func (c *Controller) del(path string, data *[]byte, ret interface{}) error {
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

func (c *Controller) handleResponse(method, path string, resp *http.Response, ret interface{}) error {
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

func (c *Controller) Initialize() (*Response, error) {
	var ret Response
	err := c.post("/initialize/", nil, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Controller) Reset() (*Response, error) {
	var ret Response
	err := c.post("/reset/", nil, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Controller) GetStatus() (*StatusResponse, error) {
	var ret StatusResponse
	err := c.get("/status/", &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Controller) GetUptime() (*Response, error) {
	var ret Response
	err := c.get("/uptime/", &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Controller) RequestShutdown() (*Response, error) {
	var ret Response
	err := c.post("/shutdown/", nil, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Controller) GetUsers() (*UsersResponse, error) {
	var ret UsersResponse
	err := c.get("/users/", &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Controller) GetBooksAdmin(username string) (*AdminBooksResponse, error) {
	var ret AdminBooksResponse
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

func (c *Controller) AddUser(username, display, password string) (*AddUserResponse, error) {
	var err error
	if display == "" {
		display = username
	}
	if password == "" {
		password, err = mkpasswd(PASSWORD_LENGTH)
		if err != nil {
			return nil, util.Fatalf("failed generating password: %v", err)
		}
	}
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
	err = c.SetPassword(username, password)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Controller) AddBook(email, name, description string) (*AddBookResponse, error) {
	if description == "" {
		description = name
	}
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

func (c *Controller) DeleteUser(username string) (*Response, error) {
	user := map[string]string{
		"username": username,
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
	err = c.DeletePassword(username)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Controller) DeleteBook(username, bookname string) (*Response, error) {
	token := util.BookToken(username, bookname)
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

func (c *Controller) Addresses(username, bookname string) (*AddressesResponse, error) {
	dav, err := c.davClient(username)
	if err != nil {
		return nil, err
	}

	addrs, err := dav.Addresses(bookname)
	if err != nil {
		return nil, err
	}
	response := AddressesResponse{}
	response.Success = true
	response.Request = "CardDAV address query"
	response.Message = fmt.Sprintf("%s %s addresses", username, bookname)
	response.Addresses = *addrs
	return &response, nil
}

func (c *Controller) GetBooks(username string) (*BooksResponse, error) {
	dav, err := c.davClient(username)
	if err != nil {
		return nil, err
	}
	books, err := dav.List()
	if err != nil {
		return nil, err
	}
	response := BooksResponse{}
	response.Success = true
	response.Request = "CardDAV address books query"
	response.Message = fmt.Sprintf("%s books", username)
	response.Books = *books
	return &response, nil

}

func (c *Controller) AddAddress(username, bookname, email, name string) (*AddressResponse, error) {
	dav, err := c.davClient(username)
	if err != nil {
		return nil, err
	}
	added, err := dav.AddAddress(bookname, email, name)
	if err != nil {
		return nil, err
	}
	response := AddressResponse{}
	response.Success = true
	response.Request = fmt.Sprintf("Add CardDAV address: %s", email)
	response.Message = fmt.Sprintf("added %s", email)
	response.Address = *added
	return &response, nil

}

func (c *Controller) DeleteAddress(username, bookname, email string) (*AddressesResponse, error) {
	dav, err := c.davClient(username)
	if err != nil {
		return nil, err
	}
	deleted, err := dav.DeleteAddress(bookname, email)
	if err != nil {
		return nil, err
	}
	response := AddressesResponse{}
	response.Success = true
	response.Request = fmt.Sprintf("Delete CardDAV address: %s", email)
	if len(*deleted) == 0 {
		response.Message = fmt.Sprintf("not found: %s", email)
	} else {
		response.Message = fmt.Sprintf("deleted: %d", len(*deleted))
	}
	response.Addresses = *deleted
	return &response, nil
}
func (c *Controller) QueryAddress(username, bookname, email string) (*AddressesResponse, error) {
	dav, err := c.davClient(username)
	if err != nil {
		return nil, err
	}
	found, err := dav.QueryAddress(bookname, email)
	if err != nil {
		return nil, err
	}
	response := AddressesResponse{}
	response.Success = true
	response.Request = fmt.Sprintf("Query CardDAV address: %s", email)
	if len(*found) == 0 {
		response.Message = fmt.Sprintf("not found: %s", email)
	} else {
		response.Message = fmt.Sprintf("found: %d", len(*found))
	}
	response.Addresses = *found
	return &response, nil
}

// return books containing address
func (c *Controller) ScanAddress(username, email string) (*BooksResponse, error) {
	dav, err := c.davClient(username)
	if err != nil {
		return nil, err
	}
	books, err := dav.ScanAddress(email)
	if err != nil {
		return nil, err
	}
	response := BooksResponse{}
	response.Success = true
	response.Request = fmt.Sprintf("Scan books for CardDAV address: %s", email)
	if len(*books) == 0 {
		response.Message = fmt.Sprintf("not found: %s", email)
	} else {
		response.Message = fmt.Sprintf("found: %d", len(*books))
	}
	response.Books = *books
	return &response, nil
}
