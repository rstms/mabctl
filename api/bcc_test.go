package api

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"testing"
)

func initConfig(t *testing.T) {
	viper.SetConfigType("yaml")
	viper.SetConfigFile("/etc/mabctl/config")
	err := viper.ReadInConfig()
	require.Nil(t, err)
	viper.Set("verbose", true)
}

func TestApiInit(t *testing.T) {

	initConfig(t)
	api, err := NewAddressBookController()
	require.Nil(t, err)

	fmt.Printf("%v\n", api)
}

func TestUsers(t *testing.T) {

	initConfig(t)
	api, err := NewAddressBookController()
	require.Nil(t, err)

	response, err := api.GetUsers()
	require.Nil(t, err)

	fmt.Printf("success: %v\n", response.Success)
	fmt.Printf("message: %v\n", response.Message)
	fmt.Printf("request: %v\n", response.Request)
	for i, user := range response.Users {
		fmt.Printf("user[%d]: %v\n", i, user)
	}
}

func TestScanBadUser(t *testing.T) {

	initConfig(t)
	api, err := NewAddressBookController()
	require.Nil(t, err)

	response, err := api.ScanAddress("username@example.org", "fromaddress@example.org")
	require.Nil(t, err)
	fmt.Printf("success: %v\n", response.Success)
	fmt.Printf("message: %v\n", response.Message)
	fmt.Printf("request: %v\n", response.Request)
	fmt.Printf("books: %v\n", response.Books)
}
