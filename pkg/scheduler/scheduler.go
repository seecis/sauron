// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package scheduler

import (
	"github.com/seecis/sauron/pkg/extractor"
)

type ExtractionScheduler interface {
	Schedule(extractor extractor.Extractor, payload ExtractionRequest) (string, error)
}

type ExtractionRequest struct {
	Url string `json:"url" yaml:"url"`
}







