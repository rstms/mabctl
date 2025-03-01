package api

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/spf13/viper"
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
    fmt.Printf("err: %v\n", err)
    fmt.Printf("response: %v\n", response)
    //fmt.Printf("success: %v\n", response.Success)
    //fmt.Printf("message: %v\n", response.Message)
    //fmt.Printf("request: %v\n", response.Request)
}

/*
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			log.Printf("BEGIN_TEST: %s\n", c.Name)
			input, err := os.ReadFile("testdata/" + c.Name)
			require.Nil(t, err)
			ibuf := bytes.NewBuffer(input)
			obuf := bytes.Buffer{}
			ebuf := bytes.Buffer{}
			cmd := exec.Command("./filterctl", "--config", "testdata/config.yml")
			cmd.Stdin = ibuf
			cmd.Stdout = &obuf
			cmd.Stderr = &ebuf
			fmt.Printf("Run %s %+v\n", c.Name, cmd)
			runErr := cmd.Run()
			var exitCode int
			if runErr != nil {
				switch e := runErr.(type) {
				case *exec.ExitError:
					exitCode = e.ExitCode()
				default:
					require.Nil(t, runErr)
				}
			} else {
				exitCode = cmd.ProcessState.ExitCode()
				:}
			fmt.Printf("Run %s returned: exitCode=%v err=%+v\n", c.Name, exitCode, runErr)
			err = os.WriteFile("testdata/"+c.Name+".out", obuf.Bytes(), 0660)
			require.Nil(t, err)
			err = os.WriteFile("testdata/"+c.Name+".err", ebuf.Bytes(), 0660)
			require.Nil(t, err)
			if c.ExpectSuccess {
				require.Zero(t, exitCode)
			} else {
				require.NotZero(t, exitCode)
			}
		})
	}
}
*/
