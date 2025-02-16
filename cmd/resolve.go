package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var ResolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Resolve a shortened URL to the original URL",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Shortened URL is required")
			return
		}

		shortURL := args[0]
		resp, err := http.Get(os.Getenv("API_URL") + "/" + shortURL)
		if err != nil {
			fmt.Println("Failed to resolve URL:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMovedPermanently {
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("Error:", string(body))
			return
		}

		fmt.Println("Original URL:", resp.Header.Get("Location"))
	},
}
