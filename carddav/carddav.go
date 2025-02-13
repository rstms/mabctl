package carddav

import (
	"context"
	"fmt"
	"github.com/emersion/go-vcard"
	"github.com/emersion/go-webdav/carddav"
	"github.com/google/uuid"
	"github.com/rstms/mabctl/util"
	"github.com/studio-b12/gowebdav"
	"net/http"
	"strings"
)

const VCARD_VERSION = "3.0"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type DigestAuthorizedClient struct {
	c        HTTPClient
	username string
	password string
	auth     gowebdav.Authenticator
}

func (c *DigestAuthorizedClient) client() *http.Client {
	ret, ok := c.c.(*http.Client)
	if !ok {
		panic("type assertion failed")
	}
	return ret
}

func (c *DigestAuthorizedClient) Do(req *http.Request) (*http.Response, error) {
	client := c.client()
	if c.auth != nil {
		err := c.auth.Authorize(client, req, req.URL.Path)
		if err != nil {
			return nil, fmt.Errorf("DigestAuthClient: preauth: %v", err)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("DigestAuthClient: Do: %v", err)
	}
	if resp.StatusCode == http.StatusUnauthorized {
		var err error
		c.auth, err = gowebdav.NewDigestAuth(c.username, c.password, resp)
		if err != nil {
			return nil, fmt.Errorf("DigestAuthClient: digest auth create: %v", err)
		}
		err = c.auth.Authorize(client, req, req.URL.Path)
		if err != nil {
			return nil, fmt.Errorf("DigestAuthClient: postauth: %v", err)
		}
		response, err := client.Do(req)
		if err != nil {
			return nil, util.Fatalf("postauth Do failed: %v", err)
		}
		return response, nil

	}
	return resp, nil
}

type CardClient struct {
	url      string
	username string
	client   *DigestAuthorizedClient
	dav      *carddav.Client
}

func NewClient(username, password, url string) (*CardClient, error) {
	if url == "" {
		var err error
		url, err = discover(username)
		if err != nil {
			return nil, err
		}
	}
	client := DigestAuthorizedClient{&http.Client{}, username, password, nil}
	dav, err := carddav.NewClient(&client, url)
	if err != nil {
		return nil, util.Fatalf("failed creating webdav client: %v", err)
	}
	c := CardClient{url, username, &client, dav}
	err = c.dav.HasSupport(context.Background())
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func discover(username string) (string, error) {

	fields := strings.Split(username, "@")
	if len(fields) != 2 {
		return "", util.Fatalf("invalid email address format: %s", username)
	}
	domain := fields[1]
	url, err := carddav.DiscoverContextURL(context.Background(), domain)
	if err != nil {
		return "", util.Fatalf("failed carddav URL discovery for domain %s :%v", domain, err)
	}
	fmt.Printf("discovered url: %s\n", url)
	return url, nil
}

func (c *CardClient) List() (*[]carddav.AddressBook, error) {
	ctx := context.Background()
	cup, err := c.dav.FindCurrentUserPrincipal(ctx)
	if err != nil {
		return nil, util.Fatalf("FindCurrentUserPrincipal failed: %v", err)
	}
	homeSet, err := c.dav.FindAddressBookHomeSet(ctx, cup)
	if err != nil {
		return nil, util.Fatalf("FindAddressBookHomeSet failed: %v", err)
	}
	books, err := c.dav.FindAddressBooks(ctx, homeSet)
	if err != nil {
		return nil, util.Fatalf("FindAddressBookHomeSet failed: %v", err)
	}
	return &books, nil
}

func (c *CardClient) Addresses(bookname string) (*[]carddav.AddressObject, error) {
	ctx := context.Background()
	uri := util.BookURI(c.username, bookname)
	query := carddav.AddressBookQuery{}
	addrs, err := c.dav.QueryAddressBook(ctx, uri, &query)
	if err != nil {
		return nil, util.Fatalf("QueryAddressBook failed: %v", err)
	}
	return &addrs, nil
}

func GetAddressEmail(address carddav.AddressObject) (string, error) {
	card := address.Card
	field := card.Get("EMAIL")
	if field != nil {
		return field.Value, nil
	}
	return "", util.Fatalf("null email address in %+v", address)
}

func GetAddressUUID(address carddav.AddressObject) (string, error) {
	card := address.Card
	field := card.Get("UID")
	if field != nil {
		return field.Value, nil
	}
	return "", util.Fatalf("null UUID in %+v", address)
}

func (c *CardClient) AddAddress(bookname, email, name string) (*carddav.AddressObject, error) {
	// if email is present in bookname, return existing object
	addrs, err := c.Addresses(bookname)
	if err != nil {
		return nil, err
	}
	for _, addr := range *addrs {
		addrEmail, err := GetAddressEmail(addr)
		if err != nil {
			return nil, err
		}
		if addrEmail == email {
			return &addr, nil
		}
	}
	ctx := context.Background()
	uuid := uuid.New()
	uri := util.BookURI(c.username, bookname)
	path := uri + uuid.String() + ".vcf"
	card := vcard.Card{}
	card.SetValue("EMAIL", email)
	card.SetValue("UID", uuid.String())
	card.SetValue("VERSION", VCARD_VERSION)
	firstName, lastName, found := strings.Cut(name, " ")
	nameField := vcard.Name{}
	if found {
		nameField.GivenName = firstName
		nameField.FamilyName = lastName
	} else {
		nameField.AdditionalName = name
	}
	card.SetName(&nameField)
	_, err = c.dav.PutAddressObject(ctx, path, card)
	if err != nil {
		return nil, err
	}
	query := carddav.AddressBookQuery{
		PropFilters: []carddav.PropFilter{
			carddav.PropFilter{
				Name: "UID",
				TextMatches: []carddav.TextMatch{
					carddav.TextMatch{
						Text: uuid.String(),
					},
				},
			},
		},
	}
	created, err := c.dav.QueryAddressBook(ctx, uri, &query)
	if err != nil {
		return nil, err
	}
	if len(created) != 1 {
		return nil, util.Fatalf("unexpected post-add query result: %+v", created)
	}
	return &created[0], nil
}

func (c *CardClient) DeleteAddress(bookname, email string) (*[]carddav.AddressObject, error) {
	ctx := context.Background()
	uri := util.BookURI(c.username, bookname)
	addrs, err := c.QueryAddress(bookname, email)
	if err != nil {
		return nil, err
	}
	for _, addr := range *addrs {
		uuid, err := GetAddressUUID(addr)
		if err != nil {
			return nil, err
		}
		path := uri + uuid + ".vcf"
		err = c.dav.RemoveAll(ctx, path)
		if err != nil {
			return nil, err
		}
	}
	return addrs, nil
}

func (c *CardClient) QueryAddress(bookname, email string) (*[]carddav.AddressObject, error) {
	ctx := context.Background()
	uri := util.BookURI(c.username, bookname)
	query := carddav.AddressBookQuery{
		PropFilters: []carddav.PropFilter{
			carddav.PropFilter{
				Name: "EMAIL",
				TextMatches: []carddav.TextMatch{
					carddav.TextMatch{
						Text: email,
					},
				},
			},
		},
	}
	addrs, err := c.dav.QueryAddressBook(ctx, uri, &query)
	if err != nil {
		return nil, err
	}
	return &addrs, nil
}

func (c *CardClient) ScanAddress(email string) (*[]carddav.AddressBook, error) {
	result := []carddav.AddressBook{}
	books, err := c.List()
	if err != nil {
		return nil, err
	}
	for _, book := range *books {

		_, bookname, _, err := util.ParseBookPath(book.Path)
		if err != nil {
			return nil, err
		}
		addrs, err := c.QueryAddress(bookname, email)
		if err != nil {
			return nil, err
		}
		if len(*addrs) > 0 {
			result = append(result, book)
		}

	}
	return &result, nil
}
