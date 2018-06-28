// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/seecis/sauron/internal/http"
)

var serverAddress string

var frontendPort string
var frontendIp string

// frontendCmd represents the web command
var frontendCmd = &cobra.Command{
	Use:   "frontend",
	Short: "Starts a html frontend server for sauron",
	Long:  `This command needs a working sauron backend to work.`,
	Run: func(cmd *cobra.Command, args []string) {
		http.ServeWeb(frontendIp, frontendPort)
	},
}

func init() {
	serveCmd.AddCommand(frontendCmd)

	frontendCmd.Flags().StringVarP(&serverAddress,
		"serverAddress",
		"",
		"sauron-backend:9091",
		"Provide address for sauron frontend to connect to sauron backend")

	frontendCmd.Flags().StringVarP(&frontendPort, "port",
		"p",
		"9090",
		"Port to bind for frontend")

	frontendCmd.Flags().StringVarP(&frontendIp,
		"ip",
		"i",
		"127.0.0.1",
		"Ip will listen to this ip")
}
