// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
)

var extractorCmd = &cobra.Command{
	Use:   "extractor",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(extractorCmd)
}
