// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.


package extractor

import "bufio"

type Field struct {
	Label     string  `json:"label" yaml:"label"`
	Data      string  `json:"data" yaml:"data"`
	Subfields []Field `json:"subfields" yaml:"subfields"`
}

type Extractor interface {
	Extract(reader *bufio.Reader) (*Field, error)
	GetName() string
}
