// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package dataaccess

import (
	"testing"
	"github.com/seecis/sauron/pkg/extractor"
	"github.com/segmentio/ksuid"
	"fmt"
	"reflect"
	"github.com/davecgh/go-spew/spew"
)

var testExtractor = extractor.HtmlExtractor{
	Name:    "Teasdasdasdst",
	Queries: nil,
	Uid:     ksuid.New(),
}

func TestMSSQLExtractorService_Save(t *testing.T) {
	id, err := e.Save(testExtractor)

	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	k, err := ksuid.Parse(id)
	testExtractor.Uid = k
	fmt.Println("Saved an html extractor with id", id)
}

func TestMSSQLExtractorService_Get(t *testing.T) {
	he, err := e.Get(testExtractor.Uid.String())
	if err != nil {
		fmt.Println("Can't get:", err)
		t.FailNow()
	}

	if !(reflect.DeepEqual(he, testExtractor)) {
		fmt.Println("Got")
		spew.Dump(he)
		fmt.Println("Expecting")
		spew.Dump(testExtractor)
		t.FailNow()
	}
}
