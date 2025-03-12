package api

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	davapi "github.com/rstms/mabctl/carddav"
	"github.com/rstms/mabctl/util"
	"github.com/spf13/viper"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

const Version = "1.4.5"

func mkpasswd(size int) (string, error) {
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func LookupDomain() (string, error) {
	domain := viper.GetString("domain")
	if domain == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return "", err
		}
		dot := strings.Index(hostname, ".")
		if dot == -1 {
			return "", fmt.Errorf("no domain in hostname: %s\n", hostname)
		}
		domain = hostname[dot+1:]
		viper.SetDefault("domain", domain)
	}
	return domain, nil
}

func LookupURL() (string, error) {
	url := viper.GetString("url")
	if url != "" {
		if viper.GetBool("verbose") {
			log.Printf("LookupURL.viper url=%s\n", url)
		}
		return url, nil
	}
	domain := viper.GetString("domain")
	if viper.GetBool("verbose") {
		log.Printf("LookupURL.viper domain=%s\n", domain)
	}

	_, records, err := net.LookupSRV("", "", fmt.Sprintf("_carddavs._tcp.%s", domain))
	if err != nil {
		return "", err
	}

	for _, record := range records {
		if viper.GetBool("verbose") {
			log.Printf("LookupURL: SRV record: %+v\n", *record)
		}
		url = strings.TrimSuffix(record.Target, ".")
		return url, nil
	}
	return "", fmt.Errorf("SRV lookup returned no records")
}

func SetDefaults() error {
	domain, err := LookupDomain()
	if err != nil {
		return err
	}
	if viper.GetBool("verbose") {
		log.Printf("LookupDomain: %s\n", domain)
	}

	url, err := LookupURL()
	if err != nil {
		return err
	}
	if viper.GetBool("verbose") {
		log.Printf("LookupURL: %s\n", url)
	}

	viper.SetDefault("bcc_url", fmt.Sprintf("https://%s:4443/bcc", url))
	viper.SetDefault("dav_url", fmt.Sprintf("https://%s/dav.php", url))
	viper.SetDefault("admin_username", "admin")
	viper.SetDefault("client_cert", "/etc/mabctl/mabctl.pem")
	viper.SetDefault("client_key", "/etc/mabctl/mabctl.key")
	viper.SetDefault("insecure_no_validate_server_certificate", false)

	for k, v := range viper.GetStringMap("domains." + domain) {
		if viper.GetBool("verbose") {
			log.Printf("viper domain default %s=%s\n", k, v)
		}
		viper.SetDefault(k, v)
	}
	return nil
}

func NewAddressBookController() (*Controller, error) {
	err := SetDefaults()
	if err != nil {
		return nil, util.Fatalf("failed setting config defaults: %v", err)
	}

	clientCert, err := tls.LoadX509KeyPair(viper.GetString("client_cert"), viper.GetString("client_key"))
	if err != nil {
		return nil, util.Fatalf("failed loading client certificate: %v", err)
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{clientCert},
		InsecureSkipVerify: viper.GetBool("insecure_no_validate_server_certificate"),
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	c := Controller{
		viper.GetString("admin_username"),
		viper.GetString("admin_password"),
		viper.GetString("bcc_url"),
		viper.GetString("api_key"),
		client,
	}

	return &c, nil
}

func (c *Controller) davClient(username string) (*davapi.CardClient, error) {
	response, err := c.GetPassword(username)
	if err != nil {
		return nil, err
	}
	password := response.Password
	url := viper.GetString("dav_url")
	cert := viper.GetString("client_cert")
	key := viper.GetString("client_key")
	insecure := viper.GetBool("insecure_no_validate_server_certificate")
	return davapi.NewClient(username, password, url, cert, key, insecure)
}
