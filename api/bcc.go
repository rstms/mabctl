package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/emersion/go-webdav/carddav"
	"github.com/rstms/mabctl/util"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"strings"
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

type AddressesResponse struct {
	Response
	Addresses []carddav.AddressObject `json:"addresses"`
}

type AddressResponse struct {
	Response
	Address carddav.AddressObject `json:"address"`
}

type AccountResponse struct {
	Response
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserAccountsResponse struct {
	Response
	Accounts map[string]string `json:"accounts"`
}

type UserAccountsRequest struct {
	Accounts map[string]string `json:"accounts"`
}

type UserDump struct {
	Password string
	Books    map[string][]string
}

type ConfigDump struct {
	Users map[string]UserDump
}

type DumpResponse struct {
	Response
	Dump ConfigDump
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
	return &ret, nil
}

func (c *Controller) AddBook(username, bookname, description string) (*AddBookResponse, error) {
	if description == "" {
		description = bookname
	}
	book := map[string]string{
		"username":    username,
		"bookname":    bookname,
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
	response.Message = fmt.Sprintf("user %s books", username)
	response.Books = make([]Book, len(*books))
	for i, book := range *books {
		book, err := c.convertBook(&book)
		if err != nil {
			return nil, err
		}
		response.Books[i] = *book
	}
	return &response, nil
}

func (c *Controller) convertBook(davBook *carddav.AddressBook) (*Book, error) {
	username, bookname, token, err := util.ParseBookPath(davBook.Path)
	if err != nil {
		return nil, err
	}
	uriIndex := strings.Index(davBook.Path, "/addressbooks/")
	if uriIndex == -1 {
		return nil, util.Fatalf("convertBook uri parse failed: %s", davBook.Path)
	}
	uri := fmt.Sprintf("%s%s", viper.GetString("dav_url"), davBook.Path[uriIndex:])

	addressesResponse, err := c.Addresses(username, bookname)
	if err != nil {
		return nil, util.Fatalf("convertBook Addresses query failed: %v", err)
	}
	contacts := len(addressesResponse.Addresses)
	book := Book{
		UserName:    username,
		BookName:    bookname,
		Description: davBook.Description,
		Contacts:    contacts,
		Token:       token,
		URI:         uri,
	}
	return &book, nil

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
	response := BooksResponse{}
	response.Success = true
	response.Request = fmt.Sprintf("Scan books for CardDAV address: %s", email)

	dav, err := c.davClient(username)
	if err != nil {
		response.Message = fmt.Sprintf("%v", err)
		return &response, nil
	}
	books, err := dav.ScanAddress(email)
	if err != nil {
		response.Success = false
		response.Message = fmt.Sprintf("%v", err)
		return &response, nil
	}
	response.Message = fmt.Sprintf("books found: %d", len(*books))
	response.Books = make([]Book, len(*books))
	for i, davBook := range *books {
		book, err := c.convertBook(&davBook)
		if err != nil {
		    response.Success=false
		    response.Message = fmt.Sprintf("%v", err)
		    return &response, nil
		}
		response.Books[i] = *book
	}
	return &response, nil
}

func (c *Controller) GetPassword(username string) (*AccountResponse, error) {
	response := AccountResponse{}
	response.Request = fmt.Sprintf("get password: %s", username)
	err := c.get("/password/"+username+"/", &response)
	if err != nil {
		response.Message = fmt.Sprintf("%v", err)
		return &response, nil
	}
	response.Success = true
	return &response, nil
}

func (c *Controller) GetAccounts() (*UserAccountsResponse, error) {
	response := UserAccountsResponse{}
	err := c.get("/accounts/", &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Controller) SetAccounts(request *UserAccountsRequest) (*UserAccountsResponse, error) {
	response := UserAccountsResponse{}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, util.Fatalf("failed formatting set passwords request data: %v", err)
	}
	err = c.post("/accounts/", &jsonData, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Controller) Dump() (*DumpResponse, error) {
	dump := ConfigDump{Users: make(map[string]UserDump)}
	accountsResponse, err := c.GetAccounts()
	if err != nil {
		return nil, err
	}
	for username, password := range accountsResponse.Accounts {
		dump.Users[username] = UserDump{Password: password, Books: make(map[string][]string)}
		booksResponse, err := c.GetBooks(username)
		if err != nil {
			return nil, err
		}
		for _, book := range booksResponse.Books {
			addressesResponse, err := c.Addresses(username, book.BookName)
			if err != nil {
				return nil, err
			}
			dump.Users[username].Books[book.BookName] = make([]string, len(addressesResponse.Addresses))
			for i, addr := range addressesResponse.Addresses {
				email := addr.Card.Get("EMAIL")
				if email == nil {
					return nil, util.Fatalf("null EMAIL value: username=%s bookname=%s %v", username, book.BookName, i)
				}
				dump.Users[username].Books[book.BookName][i] = email.Value
			}
		}
	}
	return &DumpResponse{Dump: dump}, nil
}

func (c *Controller) Restore(dump *ConfigDump) (*Response, error) {
	for username, user := range dump.Users {
		_, err := c.AddUser(username, username, user.Password)
		if err != nil {
			return nil, util.Fatalf("failed restoring username=%s: %v", username, err)
		}
		for bookname, addresses := range user.Books {
			_, err := c.AddBook(username, bookname, "")
			if err != nil {
				return nil, util.Fatalf("failed restoring username=%s bookname=%s: %v", username, bookname, err)
			}
			for _, address := range addresses {
				_, err := c.AddAddress(username, bookname, address, "")
				if err != nil {
					return nil, util.Fatalf("failed restoring username=%s bookname=%s address=%s: %v", username, bookname, address, err)
				}
			}
		}
	}
	return &Response{Request: "restore from dump", Success: true, Message: "restored"}, nil
}

func (c *Controller) Clear() (*Response, error) {
	userResponse, err := c.GetUsers()
	if err != nil {
		return nil, util.Fatalf("failed getting users: %v", err)
	}
	for _, user := range userResponse.Users {
		_, err := c.DeleteUser(user.UserName)
		if err != nil {
			return nil, util.Fatalf("failed deleting username=%s : %v", user.UserName, err)
		}
	}
	return &Response{Request: "clear", Success: true, Message: "cleared"}, nil
}
