package api

import (
	"crypto/tls"
	"fmt"
	"github.com/rstms/mabctl/carddav"
	"github.com/rstms/mabctl/util"
	"github.com/spf13/viper"
	"net/http"
)

func (c *Controller) GetPassword(username string) (string, error) {
	password := viper.GetString("passwd." + username)
	if password == "" {
		return "", fmt.Errorf("password not found for username: %s", username)
	}
	return password, nil
}

func (c *Controller) SetPassword(username, password string) error {
	viper.Set("passwd."+username, password)
	file := viper.ConfigFileUsed()
	if file != "" {
		return viper.WriteConfigAs(file)
	}
	return nil
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

