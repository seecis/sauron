// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/seecis/sauron/internal/dataaccess"
	"log"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists available executor ids",
	Run: func(cmd *cobra.Command, args []string) {
		extractors, err := dataaccess.NewFileSystemExtractorService("extractors/").GetAll()
		if err != nil {
			log.Fatal(err)
		}

		for _, v := range extractors {
			println(v.GetName())
		}
	},
}

func init() {
	extractorCmd.AddCommand(listCmd)
}
