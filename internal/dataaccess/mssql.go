// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package dataaccess

import (
	"github.com/jinzhu/gorm"
	"github.com/seecis/sauron/pkg/extractor"
	"github.com/segmentio/ksuid"
	"time"
	"github.com/pkg/errors"
	"log"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	"github.com/spf13/viper"
	"fmt"
	"strconv"
	"github.com/davecgh/go-spew/spew"
)

var g *gorm.DB

func initGorm(debug bool) {
	dbConf := viper.GetStringMap("database")

	spew.Dump(dbConf)
	u := dbConf["username"]
	p := dbConf["password"]
	h := dbConf["host"]
	po := dbConf["port"]
	dbname := dbConf["database"]

	var err error
	if g == nil {
		g, err = gorm.Open("mssql", fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", u, p, h, po, dbname))
		if err != nil {
			log.Fatal(err)
		}
	}

	g = g.Set("gorm:auto_preload", true)
	if debug {
		g = g.Debug()
	}

	if err != nil {
		log.Fatal(err)
	}
}

type FieldArray []Field

type QueryArr []Query

type Report struct {
	ID        uint64     `json:"-" gorm:"primary_key"`
	UID       []byte     `json:"-"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-" sql:"index"`
	Field     Field
	FieldId   uint64
}

type Field struct {
	ID        uint64     `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-" sql:"index"`
	SubFields []Field    `json:"subFields,omitempty" gorm:"foreignKey:ParentId"`
	Label     string     `json:"label"`
	Data      string     `json:"data"`
	ParentId  uint64     `json:"-"`
}

type Query struct {
	ID              uint       `gorm:"primary_key"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time `sql:"index"`
	Selector        string
	Name            string
	ForEachChildren bool
	SubQueries      []Query    `gorm:"foreignKey:ParentQueryId"`
	Trim            bool
	ExtractorId     uint64
	ParentQueryId   uint64
}

type HtmlExtractor struct {
	ID            uint          `gorm:"primary_key"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time    `sql:"index"`
	Name          string
	Queries       []Query       `gorm:"foreignKey:ExtractorId"`
	MetaExtractor MetaExtractor `gorm:"polymorphic:Extractor"`
}

type MetaExtractor struct {
	ID            uint       `gorm:"primary_key"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time `sql:"index"`
	Url           string
	UID           []byte
	Name          string
	ExtractorType string
	ExtractorId   uint64
}

type MSSQLExtractorService struct {
	g *gorm.DB
}

type MSSQLReportService struct {
	g *gorm.DB
}

func (q *Query) FromDomainModel(e extractor.Query) *Query {
	subqueries := (&QueryArr{}).FromDomainModelSlice(e.SubQueries)
	return &Query{
		Selector:        e.Selector,
		Name:            e.Name,
		ForEachChildren: e.ForEachChildren,
		SubQueries:      subqueries,
		Trim:            false,
	}
}

func (qa *QueryArr) FromDomainModelSlice(eqs [] extractor.Query) []Query {
	var ea QueryArr

	for _, v := range eqs {
		ea = append(ea, *(&Query{}).FromDomainModel(v))
	}

	return ea
}

func (qa *QueryArr) FromDomainModel(eqs [] extractor.Query) QueryArr {
	var ea QueryArr

	for _, v := range eqs {
		ea = append(ea, *(&Query{}).FromDomainModel(v))
	}

	return ea
}

func (q *Query) ToDomainModel(g *gorm.DB) (extractor.Query, error) {
	var subQueries []extractor.Query
	arr := q.SubQueries

	if arr != nil {
		var err error
		queryArr := QueryArr(arr)
		subQueries, err = (&queryArr).ToDomainModel(g)
		if err != nil {
			return extractor.Query{}, err
		}
	}

	return extractor.Query{
		Id:              strconv.FormatUint(uint64(q.ID), 10),
		Selector:        q.Selector,
		Name:            q.Name,
		ForEachChildren: q.ForEachChildren,
		SubQueries:      subQueries,
		Trim:            q.Trim,
	}, nil
}

func (qa *QueryArr) ToDomainModel(g *gorm.DB) ([]extractor.Query, error) {
	var bqs []extractor.Query
	for _, v := range *qa {
		dm, err := v.ToDomainModel(g)
		if err != nil {
			return nil, err
		}

		bqs = append(bqs, dm)
	}

	return bqs, nil
}

func (*HtmlExtractor) FromDomainModel(e extractor.HtmlExtractor) (*HtmlExtractor, error) {
	return &HtmlExtractor{
		Name:    e.Name,
		Queries: (&QueryArr{}).FromDomainModel(e.Queries),
		MetaExtractor: MetaExtractor{
			UID:  e.Uid.Bytes(),
			Name: e.Name,
			Url:  e.Url,
		},
	}, nil
}

func (e *MetaExtractor) ToDomainModel(g *gorm.DB) (extractor.Extractor, error) {
	switch e.ExtractorType {
	case "html_extractors":
		h := HtmlExtractor{}
		err := g.Set("gorm:auto_preload", true).Find(&h, e.ExtractorId).Error
		if err != nil {
			return nil, err
		}

		qa := QueryArr(h.Queries)
		queries, err := qa.ToDomainModel(g)
		if err != nil {
			return nil, err
		}

		if queries == nil {
			queries = make([]extractor.Query, 0)
		}

		k, err := ksuid.FromBytes(e.UID[:])
		return extractor.HtmlExtractor{
			Name:    e.Name,
			Queries: queries,
			Uid:     k,
			Url:     e.Url,
		}, err
	}

	return nil, errors.New("Unknown type")
}

func (f *FieldArray) fromDomainModel(dm []extractor.Field) FieldArray {
	var arr []Field
	for _, v := range dm {
		arr = append(arr, *(&Field{}).fromDomainModel(&v))
	}

	return arr
}

func (f *Field) fromDomainModel(field *extractor.Field) *Field {
	fa := (&FieldArray{}).fromDomainModel(field.Subfields)
	ff := []Field(fa)
	if ff == nil {
		fmt.Println("Subfields are null")
	}
	return &Field{
		SubFields: ff,
		Label:     field.Label,
		Data:      field.Data,
	}
}

func (e *MetaExtractor) BeforeCreate() error {
	e.UID = ksuid.New().Bytes()
	return nil
}

func (m *MSSQLExtractorService) GetAll() ([]extractor.Extractor, error) {

	var ma []MetaExtractor

	err := m.g.Find(&ma).Error

	// todo: wrap this
	if err != nil {
		return nil, err
	}

	des := make([]extractor.Extractor, len(ma), len(ma))

	for k, v := range ma {
		dm, e := v.ToDomainModel(m.g)
		if e != nil {
			return nil, e
		}

		des[k] = dm
	}

	return des, nil

}

func (m *MSSQLExtractorService) Save(e extractor.Extractor) (string, error) {
	switch e.(type) {
	case *extractor.HtmlExtractor:

		eh := e.(*extractor.HtmlExtractor)
		return m.saveHtmlExtractor(*eh)
	}

	return "", errors.New("Unknown type")
}

func (m *MSSQLExtractorService) saveHtmlExtractor(htmlExtractor extractor.HtmlExtractor) (string, error) {
	h, err := (&HtmlExtractor{}).FromDomainModel(htmlExtractor)
	if err != nil {
		return "", err
	}

	err = m.g.Save(&h).Error
	if err != nil {
		return "", err
	}

	k, err := ksuid.FromBytes(h.MetaExtractor.UID[:])
	return k.String(), err
}

func (m *MSSQLExtractorService) Get(id string) (extractor.Extractor, error) {
	k, err := ksuid.Parse(id)
	if err != nil {
		return nil, &DataServiceError{
			UnderlyingError: err,
			ErrorType:       BadRequestData,
			ShouldPanic:     false,
		}
	}

	me := MetaExtractor{}
	err = m.g.First(&me, "uid = ?", k.Bytes()).Error
	if err != nil {
		dse := &DataServiceError{
			UnderlyingError: err,
			ErrorType:       Unknown,
			ShouldPanic:     false,
		}

		if gorm.IsRecordNotFoundError(err) {
			dse.ErrorType = FileNotFound
			return nil, dse
		}

		return nil, dse
	}

	return me.ToDomainModel(m.g)
}

func (m *MSSQLExtractorService) Delete(id string) error {
	return errors.New("Not Implemented")
}

func (m *MSSQLExtractorService) Migrate() error {
	return m.g.AutoMigrate(&MetaExtractor{}, &HtmlExtractor{}, &Query{}).Error
}

func (m *MSSQLReportService) Get(id string) (*Report, error) {
	var r Report
	k, err := ksuid.Parse(id)
	if err != nil {
		return nil, err
	}

	err = m.g.Set("gorm:auto_preload", false).Find(&r, "uid = ?", k.Bytes()).Error
	return &r, err
}

func (r *MSSQLReportService) GetHeaders() ([]Report, error) {
	var rs []Report
	err := r.g.Set("gorm:auto_preload", false).Find(&rs).Error
	return rs, err
}

func (r *MSSQLReportService) WriteAsReport(reportId string, field *extractor.Field) error {
	k, err := ksuid.Parse(reportId)
	if err != nil {
		return &DataServiceError{
			UnderlyingError: err,
			ErrorType:       BadRequestData,
			ShouldPanic:     false,
		}
	}

	f := (&Field{}).fromDomainModel(field)
	report := Report{
		UID:   k.Bytes(),
		Field: *f,
	}

	err = r.g.Save(&report).Error
	if err != nil {
		dse := &DataServiceError{
			UnderlyingError: err,
			ErrorType:       Unknown,
			ShouldPanic:     false,
		}

		if gorm.IsRecordNotFoundError(err) {
			dse.ErrorType = FileNotFound
			return dse
		}

		return err
	}

	return nil
}

func (r *MSSQLReportService) Migrate() error {
	return r.g.AutoMigrate(&Report{}, &Field{}).Error
}

func (r *MSSQLReportService) GetAllUnjoined() ([]Report, error) {

	var reports []Report

	err := r.g.Find(&reports).Limit(10).Error
	return reports, err
}
func (r *MSSQLReportService) GetReportDetail(id string) (*Report, error) {
	rep := Report{}
	u, err := ksuid.Parse(id)
	if err != nil {
		return nil, err
	}

	err = r.g.Set("gorm:auto_preload", true).First(&rep, "uid = ?", u.Bytes()).Error

	if err != nil {
		return nil, err
	}

	return &rep, nil
}

func NewMSSQLReportService(migrate, debug bool) *MSSQLReportService {
	initGorm(debug)
	m := &MSSQLReportService{g: g}

	if !migrate {
		return m
	}

	err := m.Migrate()
	if err != nil {
		log.Fatal()
	}

	return m
}

// todo: add config.
func NewMsSqlExtractorService(migrate bool, debug bool) (m *MSSQLExtractorService) {
	initGorm(debug)
	m = &MSSQLExtractorService{g: g}
	if !migrate {
		return
	}

	err := m.Migrate()
	if err != nil {
		log.Fatal()
	}

	return
}
