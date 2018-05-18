// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package dataaccess

import (
	"os"
	"github.com/jinzhu/gorm"
	"testing"
	"log"
)

import _ "github.com/jinzhu/gorm/dialects/mssql"

var (
	e ExtractorService
	g *gorm.DB
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func shutdown() {
	g.Close()
}

func setup() {
	var err error
	g, err = gorm.Open("mssql", "sqlserver://sa:AAABBBccc123@192.168.99.102:1433?database=sauron")
	if err != nil {
		log.Fatal(err)
	}

	b := &MSSQLExtractorService{g: g.Debug()}
	err = b.Migrate()
	if err != nil {
		log.Fatal()
	}

	e = b
}
