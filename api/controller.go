package api

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"time"
	davapi "github.com/rstms/mabctl/carddav"
	"github.com/rstms/mabctl/util"
	"github.com/spf13/viper"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

const Version = "1.5.16"

func mkpasswd(size int) (string, error) {
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func LookupDomain() (string, error) {
	domain := viper.GetString("mabctl.domain")
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
		viper.SetDefault("mabctl.domain", domain)
	}
	return domain, nil
}

func LookupURL() (string, error) {
	verbose := viper.GetBool("verbose")
	url := viper.GetString("mabctl.url")
	if url != "" {
		if verbose {
		    log.Printf("LookupURL using viper mabct.url=%s\n", url)
		}
		return url, nil
	}
	domain := viper.GetString("mabctl.domain")
	if verbose {
		log.Printf("LookupURL using viper mabctl.domain=%s\n", domain)
	}

	_, records, err := net.LookupSRV("", "", fmt.Sprintf("_carddavs._tcp.%s", domain))
	if err != nil {
		return "", err
	}

	for _, record := range records {
		if verbose {
			log.Printf("LookupURL: SRV record: %+v\n", *record)
		}
		url = strings.TrimSuffix(record.Target, ".")
		return url, nil
	}
	return "", fmt.Errorf("SRV lookup returned no records")
}

func SetDefaults() error {
	verbose := viper.GetBool("verbose")
	domain, err := LookupDomain()
	if err != nil {
		return err
	}
	if verbose {
		log.Printf("LookupDomain: %s\n", domain)
	}

	url, err := LookupURL()
	if err != nil {
		return err
	}
	if verbose {
		log.Printf("LookupURL: %s\n", url)
	}

	viper.SetDefault("mabctl.bcc_url", fmt.Sprintf("https://%s:4443/bcc", url))
	viper.SetDefault("mabctl.dav_url", fmt.Sprintf("https://%s/dav.php", url))
	viper.SetDefault("mabctl.admin_username", "admin")
	viper.SetDefault("mabctl.client_cert", "/etc/mabctl/mabctl.pem")
	viper.SetDefault("mabctl.client_key", "/etc/mabctl/mabctl.key")
	viper.SetDefault("mabctl.insecure_no_validate_server_certificate", false)

	for k, v := range viper.GetStringMap("mabctl.domains." + domain) {
		if verbose {
			logValue := v.(string)
			if k == "api_key" || k == "admin_password" {
			    logValue = logValue[:4] + ".........."
			}
			log.Printf("viper domain default %s=%s\n", k, logValue)
		}
		viper.SetDefault("mabctl." + k, v)
	}
	return nil
}

func NewAddressBookController() (*Controller, error) {
	err := SetDefaults()
	if err != nil {
		return nil, util.Fatalf("failed setting config defaults: %v", err)
	}

	clientCert, err := tls.LoadX509KeyPair(viper.GetString("mabctl.client_cert"), viper.GetString("mabctl.client_key"))
	if err != nil {
		return nil, util.Fatalf("failed loading client certificate: %v", err)
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{clientCert},
		InsecureSkipVerify: viper.GetBool("mabctl.insecure_no_validate_server_certificate"),
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
			IdleConnTimeout: 5 * time.Second,
		},
	}

	c := Controller{
		viper.GetString("mabctl.admin_username"),
		viper.GetString("mabctl.admin_password"),
		viper.GetString("mabctl.bcc_url"),
		viper.GetString("mabctl.api_key"),
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
	url := viper.GetString("mabctl.dav_url")
	cert := viper.GetString("mabctl.client_cert")
	key := viper.GetString("mabctl.client_key")
	insecure := viper.GetBool("mabctl.insecure_no_validate_server_certificate")
	return davapi.NewClient(username, password, url, cert, key, insecure)
}
