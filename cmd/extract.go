package cmd

import (
	"github.com/spf13/cobra"
	"bitbucket.org/seecis/sauron/pkg/extractor"
	"os"
	"log"
	"bufio"
	"gopkg.in/yaml.v2"
	"encoding/json"
)

var file string
var extractorFile string
var outputFormat string

// extractCmd represents the extract command
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

		if outputFormat == "yaml" {
			yaml.NewEncoder(os.Stdout).Encode(*fields)
			return
		}

		if outputFormat == "json"{
			json.NewEncoder(os.Stdout).Encode(*fields)
		}

	},
}

func init() {

	rootCmd.AddCommand(extractCmd)
	extractCmd.Flags().StringVarP(&file, "file", "f", "", "Provide file to be extracted")
	extractCmd.Flags().StringVarP(&extractorFile, "extractor", "e", "", "Extractor to be used")
	extractCmd.Flags().StringVarP(&outputFormat, "outputFormat", "o", "yaml", "Output format: yaml, json")

	extractCmd.MarkFlagFilename("extractor", "yml", "yaml")
	extractCmd.MarkFlagRequired("extractor")
	extractCmd.MarkFlagRequired("file")
}
