package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

// shortenCmd represents the shorten command
var shortenCmd = &cobra.Command{
	Use:   "shorten",
	Short: "Shorten a given URL",
	Run: func(cmd *cobra.Command, args []string) {
		url, _ := cmd.Flags().GetString("url")
		customShort, _ := cmd.Flags().GetString("short")
		expiry, _ := cmd.Flags().GetInt("expiry")

		if url == "" {
			fmt.Println("Error: URL is required")
			os.Exit(1)
		}

		// Prepare request payload
		payload := map[string]interface{}{
			"url":    url,
			"short":  customShort,
			"expiry": expiry,
		}

		jsonData, _ := json.Marshal(payload)

		// Make API request
		resp, err := http.Post("http://localhost:3000/api/v1", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error: Unable to shorten URL")
			os.Exit(1)
		}
		defer resp.Body.Close()

		// Read response
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		output, _ := json.MarshalIndent(result, "", "    ")
		fmt.Println(string(output))
	},
}

func init() {
	rootCmd.AddCommand(shortenCmd)
	shortenCmd.Flags().String("url", "", "URL to shorten")
	shortenCmd.Flags().String("short", "", "Custom short link (optional)")
	shortenCmd.Flags().Int("expiry", 24, "Expiry time in hours (default: 24)")
}
