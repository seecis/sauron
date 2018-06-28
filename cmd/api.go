// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/seecis/sauron/internal/http"
)

// apiCmd represents the api command
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Serves sauron api",
	Run: func(cmd *cobra.Command, args []string) {
		http.ServeApi(backendIp, backendPort)
	},
}

var backendPort string
var backendIp string

func init() {
	serveCmd.AddCommand(apiCmd)
	apiCmd.Flags().StringVarP(&backendPort, "port", "p", "9091", "Port to bind for frontend")
	apiCmd.Flags().StringVarP(&backendIp, "ip", "i", "127.0.0.1", "Ip will listen to this ip")
}
