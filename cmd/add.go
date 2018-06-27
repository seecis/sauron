// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
	"path/filepath"
	"log"
	"os"
	"bufio"
	"github.com/seecis/sauron/pkg/extractor"
	"gopkg.in/yaml.v2"
	"github.com/seecis/sauron/cmd/util"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds extractor using yaml or json",
	Run: func(cmd *cobra.Command, args []string) {
		abs, err := filepath.Abs(ef)
		if err != nil {
			log.Fatal(err)
		}

		_, err = os.Stat(abs)
		if err != nil {
			log.Fatal(err)
		}

		file, err := os.Open(abs)
		if err != nil {
			log.Fatal(err)
		}

		defer file.Close()
		he := extractor.HtmlExtractor{}
		err = yaml.NewDecoder(bufio.NewReader(file)).Decode(&he)
		if err != nil {
			log.Fatal(err)
		}

		id, err := util.DefaultSauronHttpClient.AddExtractor(he)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(id)
	},
}

var ef string

func init() {
	extractorCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&ef, "file", "", "", "Extractor definition file")
	addCmd.MarkFlagFilename("file", "yml", "yaml")
	addCmd.MarkFlagRequired(file)
}
