// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package dataaccess

import (
	"github.com/seecis/sauron/pkg/extractor"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"os"
	"log"
)

type ExtractorService interface {
	GetAll() ([]extractor.Extractor, error)
}

// Todo: Move this code somewhere else
// Todo: Maybe put a config for this dude?
type FileSystemExtractorService struct {
}

func (fses *FileSystemExtractorService) GetAll() ([]extractor.Extractor, error) {
	files, err := ioutil.ReadDir("extractors")

	if err != nil {
		return nil, err
	}

	var extractors []extractor.Extractor
	for _, v := range files {
		file, err := os.Open("extractors/" + v.Name())
		if err != nil {
			// we wont care lel
			log.Println(err)
			continue
		}

		var tmpExtractor extractor.HtmlExtractor
		err = yaml.NewDecoder(file).Decode(&tmpExtractor)
		if err != nil {
			// we wont care lel
			log.Println(err)
			continue
		}

		extractors = append(extractors, &tmpExtractor)
	}

	return extractors, nil
}
