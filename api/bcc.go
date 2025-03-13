package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/emersion/go-webdav/carddav"
	davapi "github.com/rstms/mabctl/carddav"
	"github.com/rstms/mabctl/util"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
	User string `json:"user"`
	Message string `json:"message"`
	Request string `json:"request"`
}

type UsersResponse struct {
	Response
	Users []User `json:"users"`
}

type UserBooksResponse struct {
	Response
	UserBooks map[string][]string
}

type AddUserResponse struct {
	Response
	User User `json:"user"`
}

type BooksResponse struct {
	Response
	Books []Book	`json:"books"`
}

type AddBookResponse struct {
	Response
	Book Book `json:"book"`
}

type StatusResponse struct {
	Response
	Status map[string]any `json:"status"`
}

type ErrorResponse struct {
	Response
	Detail string `json:"detail"`
}

type AddressesResponse struct {
	Response
	Addresses []string `json:"addresses"`
}

type AddressResponse struct {
	Response
	Address *carddav.AddressObject `json:"address"`
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

func URIPath(uri string) (string, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return "", util.Fatalf("failed parsing URI %s: %v", uri, err)
	}
	return parsed.Path, nil
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

func (c *Controller) GetUserBooks() (*UserBooksResponse, error) {
	var ret UserBooksResponse
	ret.Success = true
	ret.Message = "all users and address books"
	ret.Request = "get user books"
	usersResponse, err := c.GetUsers()
	if err != nil {
		return nil, err
	}
	ret.UserBooks = make(map[string][]string)
	for _, user := range usersResponse.Users {
		booksResponse, err := c.GetBooks(user.UserName)
		if err != nil {
			return nil, err
		}
		books := make([]string, len(booksResponse.Books))
		for i, book := range booksResponse.Books {
		    books[i] = book.BookName
		}
	    	ret.UserBooks[user.UserName] = books
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
	_, err = c.DeleteBook(username, "default address book")
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
	booksResponse, err := c.GetBooks(username)
	if err != nil {
	    return nil, util.Fatalf("requesting existing books: %v", err)
	}
	for _, book := range booksResponse.Books {
	    if book.BookName == bookname {
		var response AddBookResponse
		response.Success = true
		response.Request = "add book"
		response.Message = "book Exists"
		response.Book = book
		return &response, nil
	    }
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

func (c *Controller) Addresses(dav *davapi.CardClient, username, bookname string) (*AddressesResponse, error) {

	if dav == nil {
		var err error
		dav, err = c.davClient(username)
		if err != nil {
			return nil, err
		}
	}
	book, err := c.GetBook(username, bookname)
	if err != nil {
		return nil, err
	}
	path, err := URIPath(book.URI)
	if err != nil {
		return nil, err
	}
	davAddrs, err := dav.Addresses(path)
	if err != nil {
		return nil, err
	}

	addrs, err := c.EmailAddressList(davAddrs)
	if err != nil {
	    return nil, err
	}

	response := AddressesResponse{}
	response.Success = true
	response.Request = "address book addresses"
	response.Message = fmt.Sprintf("%s %s addresses", username, bookname)
	response.Addresses = *addrs

	return &response, nil
}

func (c *Controller) GetBook(username, bookname string) (*Book, error) {
	response, err := c.GetBooks(username)
	if err != nil {
		return nil, err
	}
	for _, book := range response.Books {
		if book.BookName == bookname {
			return &book, nil
		}
	}
	return nil, util.Fatalf("book name not found: %s", bookname)
}

func (c *Controller) GetBooks(username string) (*BooksResponse, error) {
	response := BooksResponse{}
	err := c.get(fmt.Sprintf("/books/%s/", username), &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Controller) convertBook(dav *davapi.CardClient, davBook *carddav.AddressBook, detailed bool) (*Book, error) {
	username, bookname, token, err := util.ParseBookPath(davBook.Path)
	if err != nil {
		return nil, err
	}
	book := Book{
		UserName:    username,
		BookName:    bookname,
		Description: davBook.Description,
		Token:       token,
	}
	if detailed {
		uriIndex := strings.Index(davBook.Path, "/addressbooks/")
		if uriIndex == -1 {
			return nil, util.Fatalf("convertBook uri parse failed: %s", davBook.Path)
		}
		book.URI = fmt.Sprintf("%s%s", viper.GetString("mabctl.dav_url"), davBook.Path[uriIndex:])

		addressesResponse, err := c.Addresses(dav, username, bookname)
		if err != nil {
			return nil, util.Fatalf("convertBook Addresses query failed: %v", err)
		}
		book.Contacts = len(addressesResponse.Addresses)
	}
	return &book, nil

}

func (c *Controller) AddAddress(dav *davapi.CardClient, username, bookname, email, name string) (*AddressResponse, error) {
	verbose := viper.GetBool("verbose")
	var err error
	if dav == nil {
		dav, err = c.davClient(username)
		if err != nil {
			return nil, err
		}
	}
	response := AddressResponse{}
	response.Success = true
	response.Request = fmt.Sprintf("Add CardDAV address: %s", email)
	found, err := dav.QueryAddress(bookname, email)
	if err != nil {
		return nil, err
	}
	for _, addr := range *found {
	    if verbose {
		log.Printf("AddAddress: found existing: %+v\n", addr)
	    }
	    response.Address = &addr
	    response.Message = fmt.Sprintf("existing %s", email)
	    return &response, nil
	}

	    added, err := dav.AddAddress(bookname, email, name)
	    if err != nil {
		return nil, err
	    }
	    response.Address = added
	    response.Message = fmt.Sprintf("added %s", email)
	    return &response, nil

}

func (c *Controller) EmailAddress(addr carddav.AddressObject) (string, error) {
	    field := addr.Card.Get("EMAIL")
	    if field == nil {
		return "", util.Fatalf("cardDAV address EMAIL lookup failed: %+v", addr)
	    }
	    return field.Value, nil
}

func (c *Controller) EmailAddressList(addrs *[]carddav.AddressObject) (*[]string, error) {
	ret := []string{}
	for _, addr := range *addrs {
	    email, err := c.EmailAddress(addr)
	    if err != nil {
		return nil, err
	    }

	    ret = append(ret, email)
	}
	return &ret, nil
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
	addrs, err := c.EmailAddressList(deleted)
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
	response.Addresses = *addrs 
	return &response, nil
}

func (c *Controller) QueryAddress(username, bookname, email string) (*AddressResponse, error) {
	dav, err := c.davClient(username)
	if err != nil {
		return nil, err
	}
	found, err := dav.QueryAddress(bookname, email)
	if err != nil {
		return nil, err
	    }
	
	response := AddressResponse{}
	response.Success = true
	response.Request = fmt.Sprintf("Query CardDAV address: %s", email)
	if len(*found) == 0 {
		response.Message = fmt.Sprintf("not found: %s", email)
		response.Address = nil
	} else {
		response.Message = fmt.Sprintf("found: %d", len(*found))
		response.Address = &(*found)[0]
	}
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
		book, err := c.convertBook(dav, &davBook, false)
		if err != nil {
			response.Success = false
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

func (c *Controller) Dump(dumpUser string) (*DumpResponse, error) {
	verbose := viper.GetBool("verbose")
	dump := ConfigDump{Users: make(map[string]UserDump)}
	usersResponse, err := c.GetUsers()
	if err != nil {
		return nil, err
	}
	type DumpResult struct {
		username string
		err      error
		ret      UserDump
	}
	results := make(chan DumpResult)
	accountsResponse, err := c.GetAccounts()
	if err != nil {
	    return nil, err
	}
	resultCount := 0
	for _, user := range usersResponse.Users {
		if dumpUser != "" && user.UserName != dumpUser {
		    continue
		}
		resultCount++
		go func(username string, accounts map[string]string) {
			if verbose {
				log.Printf("dumping user %s\n", username)
			}
			var dav *davapi.CardClient
			password, ok := accounts[username]
			if !ok {
			    results <- DumpResult{username, fmt.Errorf("password not found: username=%s", username), UserDump{}}
				return
			}
			userdump := UserDump{Password: password, Books: make(map[string][]string)}
			booksResponse, err := c.GetBooks(username)
			if err != nil {
				results <- DumpResult{username, err, UserDump{}}
				return
			}
			for _, book := range booksResponse.Books {
				if verbose {
					log.Printf("dumping book %s/%s\n", username, book.BookName)
				}

				if book.Contacts == 0 {
					userdump.Books[book.BookName] = []string{}
				} else {

					if dav == nil {
						d, err := c.davClient(username)
						if err != nil {
							results <- DumpResult{username, err, UserDump{}}
							return
						}
						dav = d
						if verbose {
							log.Printf("created davClient: %+v\n", dav)
						}
					}

					addressesResponse, err := c.Addresses(dav, username, book.BookName)
					if err != nil {
						results <- DumpResult{username, err, UserDump{}}
						return
					}
					userdump.Books[book.BookName] = make([]string, len(addressesResponse.Addresses))
					for i, email := range addressesResponse.Addresses {
						userdump.Books[book.BookName][i] = email
						if verbose {
							log.Printf("dumping address %s/%s/%s\n", username, book.BookName, email)
						}
					}
				}
			}
			results <- DumpResult{username, nil, userdump}
		}(user.UserName, accountsResponse.Accounts)

	}

	for i := 0; i < resultCount; i++ {
		result := <-results
		if result.err != nil {
			err = result.err
		} else {
			dump.Users[result.username] = result.ret
		}
	}
	if err != nil {
		return nil, err
	}
	var ret DumpResponse
	ret.Success = true
	ret.Request = "dump all"
	if dumpUser != "" {
	    ret.Request = fmt.Sprintf("dump user %s", dumpUser)
	}
	ret.Message = "dumped"
	ret.Dump = dump
	return &ret, nil
}

func (c *Controller) Restore(dump *ConfigDump, restoreUser string) (*Response, error) {
	verbose := viper.GetBool("verbose")
	type RestoreResult struct {
		username string
		err      error
	}
	results := make(chan RestoreResult)
	type BookAddrs struct {
		bookname  string
		addresses []string
	}

	userbooks := make(map[string][]BookAddrs)

	for username, user := range dump.Users {
		if restoreUser != "" && username != restoreUser {
		    continue
		}
		_, err := c.AddUser(username, username, user.Password)
		if err != nil {
			return nil, util.Fatalf("failed restoring username=%s: %v", username, err)
		}
		if verbose {
			log.Printf("created user: %s\n", username)
		}

		userjobs := []BookAddrs{}
		for bookname, addresses := range user.Books {
			if strings.ToLower(bookname) != "default address book" {
				_, err := c.AddBook(username, bookname, "")
				if err != nil {
					return nil, util.Fatalf("failed restoring username=%s bookname=%s: %v", username, bookname, err)
				}
				if verbose {
					log.Printf("created book: %s/%s [%d]\n", username, bookname, len(addresses))
				}
				if len(addresses) > 0 {
					userjobs = append(userjobs, BookAddrs{bookname, addresses})
				}
			}
		}
		userbooks[username] = userjobs
	}

	resultCount := 0
	for username, jobs := range userbooks {
		if restoreUser != "" && username != restoreUser {
		    continue 
		}
		resultCount++

		go func(username string, jobs []BookAddrs, results chan RestoreResult) {
			if verbose {
				log.Printf("restore[%s]: begin\n", username)
			}
			var dav *davapi.CardClient
			for _, job := range jobs {
				if verbose {
					log.Printf("restore[%s]: job=%+v\n", username, job)
				}
				for _, address := range job.addresses {

					if dav == nil {
						d, err := c.davClient(username)
						if err != nil {
							results <- RestoreResult{username, err}
							return
						}
						dav = d
						if verbose {
							log.Printf("restore[%s]: created dav client: %v\n", username, dav)
						}
					}

					_, err := c.AddAddress(dav, username, job.bookname, address, "")
					if err != nil {
						results <- RestoreResult{username, util.Fatalf("failed restoring username=%s bookname=%s address=%s: %v", username, job.bookname, address, err)}
					}
					if verbose {
						log.Printf("restore[%s]: created address: %s/%s/%s\n", username, username, job.bookname, address)
					}
				}
			}
			if verbose {
				log.Printf("restore[%s]: success\n", username)
			}
			results <- RestoreResult{username, nil}
		}(username, jobs, results)
	}

	errors := []string{}

	for i := 0; i < resultCount; i++ {
		result := <-results
		if result.err != nil {
			errors = append(errors, fmt.Sprintf("restore[%s] fail: %v", result.username, result.err))
		}
	}
	if len(errors) > 0 {
		return nil, fmt.Errorf("%s", strings.Join(errors, "\n"))
	}
	return &Response{Request: "restore", Success: true, Message: "restored"}, nil
}

func (c *Controller) Clear() (*Response, error) {
	verbose := viper.GetBool("verbose")
	users := make(map[string]bool)
	accounts := make(map[string]bool)
	names := make(map[string]bool)

	accountsResponse, err := c.GetAccounts()
	if err != nil {
		return nil, util.Fatalf("failed getting accounts: %v", err)
	}
	for username, _ := range accountsResponse.Accounts {
		accounts[username] = true
		names[username] = true
	}

	usersResponse, err := c.GetUsers()
	if err != nil {
		return nil, util.Fatalf("failed getting users: %v", err)
	}
	for _, user := range usersResponse.Users {
		users[user.UserName] = true
		names[user.UserName] = true
	}

	for username, _ := range names {
		_, ok := users[username]
		if ok {
			if verbose {
				log.Printf("clearing user %s\n", username)
			}
			_, err := c.DeleteUser(username)
			if err != nil {
				return nil, err
			}
			delete(accounts, username)
		}
		_, ok = accounts[username]
		if ok {
			if verbose {
				log.Printf("clearing account %s\n", username)
			}
			c.DeleteUser(username)
		}
	}
	return &Response{Request: "clear", Success: true, Message: "cleared"}, nil
}
