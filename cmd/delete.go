/*
Copyright © 2024 Matt Krueger <mkrueger@rstms.net>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete USERNAME BOOKNAME EMAIL",
	Short: "delete email adddress",
	Long: `
Delete an email address from the CardDAV address book BOOKNAME under the user
account USERNAME
`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		bookname := args[1]
		email := args[2]
		response, err := MAB.DeleteAddress(username, bookname, email)
		cobra.CheckErr(err)
		if !HandleResponse(response, response.Addresses) {
		    for _, address := range response.Addresses {
			cmd.Println(address.Path)
		    }
		    /*
		    cmd.Printf("response: %v\n", response)
		    cmd.Printf("response.Address: %v\n", response.Address)
		    cmd.Printf("response.Address.Card: %v\n", response.Address.Card)
		    cmd.Printf("response.Address.Card.Get(EMAIL): %v\n", response.Address.Card.Get("EMAIL"))
		    for k, v := range response.Address.Card {
			cmd.Printf("%v %v\n", k, v)
		    }
		    card := response.Address.Card
		    field := card.Get("UID") 
		    if field != nil {
			cmd.Printf("%s.vcf\n", field.Value)
		    } else {
			cobra.CheckErr(fmt.Errorf("unexpected response: %+v", response))
		    }
		    */
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
