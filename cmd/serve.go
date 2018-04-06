package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var webPort string
var bindIp string

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serves sauron",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("serve called")
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().StringVar(&webPort, "port", "8080", "Provide port for server")
	serveCmd.PersistentFlags().StringVar(&bindIp, "bind", "0.0.0.0", "IP to bind")
}
