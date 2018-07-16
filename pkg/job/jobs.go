// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package job

import (
	"github.com/segmentio/ksuid"
	"time"
)

type ReportHeader struct {
	UID     ksuid.KSUID `json:"id"`
	Status  string      `json:"status"`
	Message string      `json:"message"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}
