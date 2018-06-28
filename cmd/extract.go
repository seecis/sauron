// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/seecis/sauron/pkg/extractor"
	"os"
	"log"
	"bufio"
	"gopkg.in/yaml.v2"
	"encoding/json"
)

var file string
var extractorFile string
var outputFormat string

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extracts data using given extractor from a file",
	Long:  `Extracts data using given extractor from a file`,
	Run: func(cmd *cobra.Command, args []string) {
		yamlFile, err := os.Open(extractorFile)
		if err != nil {
			log.Fatal(err)
		}

		ff, err := os.Open(file)
		ex := extractor.NewHtmlExtractor(bufio.NewReader(yamlFile))
		fields, err := ex.Extract(bufio.NewReader(ff))

		if err != nil {
			log.Fatal(err)
			return
		}

		switch outputFormat {
		case "yaml":
			err = yaml.NewEncoder(os.Stdout).Encode(*fields)
		case "json":
			err = json.NewEncoder(os.Stdout).Encode(*fields)
		}

		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(extractCmd)
	extractCmd.Flags().StringVarP(&file, "file", "f", "", "File which extractor will be run on")
	extractCmd.Flags().StringVarP(&extractorFile, "extractor", "e", "", "Extractor id")
	extractCmd.Flags().StringVarP(&outputFormat, "outputFormat", "o", "yaml", "Output format: yaml, json")

	extractCmd.MarkFlagFilename("extractor", "yml", "yaml")
	extractCmd.MarkFlagRequired("extractor")
	extractCmd.MarkFlagRequired("file")
}
