// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package scheduler

type ExtractionScheduler interface {
	Schedule(extractorId uint64, payload ExtractionRequest) (string, error)
	ScheduleSync(extractorId uint64, payload ExtractionRequest) (string, error)
}

type ExtractionRequest struct {
	Url      string `json:"url" yaml:"url"`
	ReportId uint64 `json:"-"`
}
