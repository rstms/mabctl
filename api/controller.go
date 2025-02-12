package api

import (
	"crypto/tls"
	"os"
	"strings"
	"fmt"
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"github.com/rstms/mabctl/carddav"
	"github.com/rstms/mabctl/util"
	"github.com/spf13/viper"
	"net/http"
)

func mkpasswd(size int) (string, error) {
    bytes := make([]byte, size)
    _, err := rand.Read(bytes)
    if err != nil {
	return "", err
    }
    return hex.EncodeToString(bytes), nil
}

func readAccounts() (map[string]string, error) {
	viper.SetDefault("passwd", "/etc/mabctl/passwd")
	passwd := viper.GetString("passwd")
	accounts := make(map[string]string)
	_, err := os.Stat(passwd)
	if err != nil {
	    if os.IsNotExist(err) {
		return accounts, nil
	    }
	    return map[string]string{}, err
	}
	file, err := os.Open(passwd)
	if err != nil {
	    return map[string]string{}, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
	    fields := strings.Split(scanner.Text(), ":")
	    if len(fields) != 2 {
		return map[string]string{}, fmt.Errorf("failed parsing passwd file: %s", passwd)
	    }
	    accounts[fields[0]]=fields[1]
	}
	return accounts, nil
}

func writeAccounts(accounts map[string]string) error {
	viper.SetDefault("passwd", "/etc/mabctl/passwd")
	passwd := viper.GetString("passwd")
	content := ""
	for username,password := range accounts {
	    if username == "" {
		return fmt.Errorf("illegal null username")
	    }
	    if strings.Contains(username, ":") {
		return fmt.Errorf("illegal ':' in username: %s", username)
	    }
	    if password == "" {
		return fmt.Errorf("illegal null username")
	    }
	    if strings.Contains(password, ":") {
		return fmt.Errorf("illegal password character: %s", ":")
	    }
	    content += fmt.Sprintf("%s:%s\n", username, password)
	}
	return os.WriteFile(passwd, []byte(content), 0600)
}

func (c *Controller) GetPassword(username string) (string, error) {
    accounts, err := readAccounts()
    if err != nil {
	return "", err
    }
    password, ok := accounts[username]
    if ok {
	return password, nil
    }
    return "", fmt.Errorf("password not found for username: %s", username)
}


func (c *Controller) SetPassword(username, password string) error {
    accounts, err := readAccounts()
    if err != nil {
	return err
    }
    accounts[username] = password
    return writeAccounts(accounts)
}

func (c *Controller) DeletePassword(username string) error {
    accounts, err := readAccounts()
    if err != nil {
	return err
    }
    delete(accounts, username)
    return writeAccounts(accounts)
}

func NewAddressBookController() (*Controller, error) {
	username := viper.GetString("admin_username")
	password := viper.GetString("admin_password")
	url := viper.GetString("admin_url")
	cert := viper.GetString("cert")
	key := viper.GetString("key")
	apikey := viper.GetString("api_key")
	insecure := viper.GetBool("insecure")

	c := Controller{username, password, url, cert, key, apikey, nil}

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

func (c *Controller) davClient(username string) (*carddav.CardClient, error) {
	password, err := c.GetPassword(username)
	if err != nil {
		return nil, err
	}
	url := viper.GetString("dav_url")
	return carddav.NewClient(username, password, url)
}

