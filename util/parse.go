package util

import (
	"fmt"
	"strings"
	"unicode"
)

// return username, bookname from token
func ParseBookToken(username, token string) (string, string, error) {
	if token == "default" {
		return username, "Default Address Book", nil
	}

	prefix := tokenize(username) + "-"
	if strings.HasPrefix(token, prefix) {
		book := token[len(prefix):]
		return username, book, nil
	}
	return "", "", fmt.Errorf("unexpected token format: %s", token)
}


// return book token from username, bookname
func BookToken(username, bookname string) string {
	if strings.ToLower(bookname) == "default address book" {
		return "default"
	}
	return tokenize(username + "-" + bookname)
}

func tokenize(raw string) string {
	token := ""
	for _, char := range raw {
		if unicode.IsUpper(char) {
			char = unicode.ToLower(char)
		}
		if unicode.IsLower(char) || unicode.IsDigit(char) {
			token = token + string(char)
		} else {
			token = token + "-"
		}
	}
	return token
}

// return username, bookname, token from book path
func ParseBookPath(path string) (string, string, string, error) {
	fields := strings.SplitN(path, "/", -1)
	if len(fields) < 4 {
		return "", "", "", Fatalf("unexpected book path: %s", path)
	}
	username := fields[len(fields)-3]
	token := fields[len(fields)-2]
	username, bookname, err := ParseBookToken(username, token)
	if err != nil {
		return "", "", "", err
	}
	return username, bookname, token, nil
}

// return book URI given username, bookname
func BookURI(username, bookname string) string {
	token := BookToken(username, bookname)
	bookURI := fmt.Sprintf("/dav.php/addressbooks/%s/%s/", username, token)
	return bookURI
}
