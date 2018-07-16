// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package extractor

import (
	"io"
	"github.com/segmentio/ksuid"
)

type Field struct {
	Label     string  `json:"label" yaml:"label"`
	Data      string  `json:"data" yaml:"data"`
	Subfields []Field `json:"subfields" yaml:"subfields"`
}

type Extractor interface {
	Extract(reader io.Reader) (*Field, error)
	GetName() string
	GetUid() ksuid.KSUID
	GetId() uint64
}
