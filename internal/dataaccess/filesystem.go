// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package dataaccess

import (
	"github.com/seecis/sauron/pkg/extractor"
	"io/ioutil"
	"strings"
	"os"
	"gopkg.in/yaml.v2"
)

type FileSystemExtractorService struct {
	basePath string
}

// returns: if there is an error the type of it is always *DataServiceError
func (fses *FileSystemExtractorService) GetAll() ([]extractor.Extractor, error) {
	files, err := ioutil.ReadDir(fses.basePath)
	if err != nil {
		return nil, &DataServiceError{
			UnderlyingError: err,
			ErrorType:       Unknown,
			ShouldPanic:     true,
		}
	}

	var extractors []extractor.Extractor
	for _, v := range files {
		fileName := v.Name()
		extractorName := fileName[0:strings.Index(fileName, ".yaml")]
		ex, _ := fses.Get(extractorName)
		//todo: no!
		extractors = append(extractors, ex)
	}

	return extractors, nil
}

func (fses *FileSystemExtractorService) Save(e extractor.Extractor) (string, error) {
	switch e.(type) {
	case *extractor.HtmlExtractor:
		v := e.(*extractor.HtmlExtractor)
		fileName := fses.basePath + v.Name + ".yaml"
		file, err := os.Create(fileName)
		defer file.Close()
		if err != nil {
			return "", &DataServiceError{
				UnderlyingError: err,
				ErrorType:       Unknown,
				ShouldPanic:     true,
			}
		}

		yaml.NewEncoder(file).Encode(v)
		return v.Name, nil
	}

	return "", nil
}

func (fses *FileSystemExtractorService) Get(id string) (extractor.Extractor, error) {
	file, err := os.Open(fses.basePath + id + ".yaml")
	defer file.Close()
	if err != nil {
		return nil, &DataServiceError{
			UnderlyingError: err,
			ErrorType:       FileNotFound,
			ShouldPanic:     false,
		}
	}

	var h extractor.HtmlExtractor
	err = yaml.NewDecoder(file).Decode(&h)
	if err != nil {
		return nil, &DataServiceError{
			UnderlyingError: err,
			ErrorType:       MalformedLocalData,
			ShouldPanic:     true,
		}
	}

	return &h, nil
}

func (fses *FileSystemExtractorService) Delete(id string) error {
	extractors, err := ioutil.ReadDir(fses.basePath)
	if err != nil {
		return buildUnknownError(err)
	}

	var deletePath = ""
	for _, v := range extractors {
		if v.Name() == id+".yaml" {
			deletePath = fses.basePath + v.Name()
			break
		}
	}

	if deletePath == "" {
		return buildFileNotFoundErr(nil)
	}

	err = os.Remove(fses.basePath + id + ".yaml")
	if err != nil {
		return buildUnknownError(err)
	}

	return nil
}

func buildFileNotFoundErr(err error) *DataServiceError {
	return &DataServiceError{
		err,
		FileNotFound,
		false,
	}
}

func buildUnknownError(err error) *DataServiceError {
	return &DataServiceError{
		err,
		Unknown,
		false,
	}
}


