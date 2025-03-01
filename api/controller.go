package api

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"github.com/rstms/mabctl/carddav"
	"github.com/rstms/mabctl/util"
	"github.com/spf13/viper"
	"net"
	"net/http"
	"os"
	"strings"
)

const Version = "1.3.8"

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
		accounts[fields[0]] = fields[1]
	}
	return accounts, nil
}

func writeAccounts(accounts map[string]string) error {
	viper.SetDefault("passwd", "/etc/mabctl/passwd")
	passwd := viper.GetString("passwd")
	content := ""
	for username, password := range accounts {
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
	return os.WriteFile(passwd, []byte(content), 0660)
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
	}
	return domain, nil
}

func LookupURL() (string, error) {
	url := viper.GetString("url")
	if url != "" {
		return url, nil
	}
	domain, err := LookupDomain()
	if err != nil {
		return "", err
	}
	_, records, err := net.LookupSRV("", "", fmt.Sprintf("_carddavs._tcp.%s", domain))
	if err != nil {
		return "", err
	}
	for _, record := range records {
		return strings.TrimSuffix(record.Target, "."), nil
	}
	return url, fmt.Errorf("SRV lookup returned no records")
}

func SetDefaults() error {
	url, err := LookupURL()
	if err != nil {
		return err
	}
	viper.SetDefault("admin_url", fmt.Sprintf("https://%s:4443/bcc", url))
	viper.SetDefault("dav_url", fmt.Sprintf("https://%s/dav.php", url))
	viper.SetDefault("admin_username", "admin")
	viper.SetDefault("cert", "/etc/mabctl/mabctl.pem")
	viper.SetDefault("key", "/etc/mabctl/mabctl.key")
	viper.SetDefault("insecure", false)
	viper.SetDefault("passwd", "/etc/mabctl/passwd")
	return nil
}

func NewAddressBookController() (*Controller, error) {
	err := SetDefaults()
	if err != nil {
		return nil, util.Fatalf("failed setting config defaults: %v", err)
	}

	clientCert, err := tls.LoadX509KeyPair(viper.GetString("cert"), viper.GetString("key"))
	if err != nil {
		return nil, util.Fatalf("failed loading client certificate: %v", err)
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{clientCert},
		InsecureSkipVerify: viper.GetBool("insecure"),
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	c := Controller{
		viper.GetString("admin_username"),
		viper.GetString("admin_password"),
		viper.GetString("admin_url"),
		viper.GetString("api_key"),
		client,
	}

	return &c, nil
}

func (c *Controller) davClient(username string) (*carddav.CardClient, error) {
	passwordResponse, err := c.GetPassword(username)
	if err != nil {
		return nil, err
	}
	if !passwordResponse.Success {
		return nil, fmt.Errorf("Unknown user: %s", username)
	}
	url := viper.GetString("dav_url")
	cert := viper.GetString("cert")
	key := viper.GetString("key")
	insecure := viper.GetBool("insecure")
	return carddav.NewClient(username, passwordResponse.Password, url, cert, key, insecure)
}
