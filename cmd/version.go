// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
)

//Todo: Fetch this from somewhere else
var VERSION = "0.0.1"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(VERSION)
	},
}
