// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package dataaccess

import (
	"github.com/seecis/sauron/pkg/extractor"
	"fmt"
)

type errorType int

const (
	Unknown            errorType = iota
	FileNotFound
	MalformedLocalData
	BadRequestData
)

func IsNotFound(err error) bool {
	switch err.(type) {
	case *DataServiceError:
		dse := err.(*DataServiceError)
		return dse.ErrorType == FileNotFound
	default:
		return false
	}
}

func IsBadRequest(err error) bool {
	switch err.(type) {
	case *DataServiceError:
		dse := err.(*DataServiceError)
		return dse.ErrorType == BadRequestData
	default:
		return false
	}
}

type DataServiceError struct {
	UnderlyingError error
	ErrorType       errorType
	ShouldPanic     bool //Todo this shouldn't be here at all!
}

type ExtractorService interface {
	GetAll() ([]extractor.Extractor, error)
	Save(e extractor.Extractor) (string, error)
	Get(id string) (extractor.Extractor, error)
	GetInternal(id uint) (extractor.Extractor, error)
	Delete(id string) error
}

type ReportService interface {
	WriteAsReport(reportId uint64, field *extractor.Field) error
	GetHeaders() ([]Report, error)
	Get(id string) (*Report, error)
	GetAll() ([]*Report, error)
	CreateForJob(j *Job, v *Version) (*Report, error)
}

type JobService interface {
	GetAll() ([]*Job, error)
	Save(j *Job) (string, error)
	Get(id string) (*Job, error)
	Delete(id string)
}

func (e *DataServiceError) Error() string {
	return fmt.Sprintf("ShouldPanic: %t, type: %d, underlying error message: %s",
		e.ShouldPanic,
		e.ErrorType,
		e.UnderlyingError.Error())
}
