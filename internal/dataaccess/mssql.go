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
)

type QueryArr []Query

func (q *Query) FromDomainModel(e extractor.Query) *Query {
	subqueries := (&QueryArr{}).FromDomainModel(e.SubQueries)
	return &Query{
		Selector:        e.Selector,
		Name:            e.Name,
		ForEachChildren: e.ForEachChildren,
		SubQueries:      &subqueries,
		Trim:            false,
	}
}

func (q *QueryArr) FromDomainModel(eqs [] extractor.Query) QueryArr {
	var ea QueryArr

	for _, v := range eqs {
		ea = append(ea, *(&Query{}).FromDomainModel(v))
	}

	return ea
}

func (q *Query) ToDomainModel(g *gorm.DB) (extractor.Query, error) {
	subQueries, err := q.SubQueries.ToDomainModel(g)
	if err != nil {
		return extractor.Query{}, err
	}

	return extractor.Query{
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
		},
	}, nil
}

type Query struct {
	ID              uint       `gorm:"primary_key"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time `sql:"index"`
	Selector        string
	Name            string
	ForEachChildren bool
	SubQueries      *QueryArr
	Trim            bool
}

type HtmlExtractor struct {
	ID            uint          `gorm:"primary_key"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time    `sql:"index"`
	Name          string
	Queries       QueryArr
	MetaExtractor MetaExtractor `gorm:"polymorphic:Extractor;"`
}

type MetaExtractor struct {
	ID            uint       `gorm:"primary_key"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time `sql:"index"`
	UID           []byte
	Name          string
	ExtractorType string
	ExtractorId   uint64
}

func (m *MSSQLExtractorService) Migrate() error {
	return m.g.AutoMigrate(&MetaExtractor{}, &HtmlExtractor{}, &Query{}).Error
}

func (e *MetaExtractor) ToDomainModel(g *gorm.DB) (extractor.Extractor, error) {
	switch e.ExtractorType {
	case "html_extractors":
		h := HtmlExtractor{}
		err := g.Model(e).Related(&h,"ExtractorId").Error
		if err != nil {
			return nil, err
		}

		queries, err := h.Queries.ToDomainModel(g)
		if err != nil {
			return nil, err
		}

		k, err := ksuid.FromBytes(e.UID[:])
		return extractor.HtmlExtractor{
			Name:    e.Name,
			Queries: queries,
			Uid:     k,
		}, err
	}

	return nil, errors.New("Unknown type")
}

func (e *MetaExtractor) BeforeCreate() error {
	e.UID = ksuid.New().Bytes()
	return nil
}

type MSSQLExtractorService struct {
	g *gorm.DB
}

func (m *MSSQLExtractorService) GetAll() ([]extractor.Extractor, error) {

	return nil, errors.New("Not Implemented")
}

func (m *MSSQLExtractorService) Save(e extractor.Extractor) (string, error) {
	switch e.(type) {
	case extractor.HtmlExtractor:
		return m.saveHtmlExtractor(e.(extractor.HtmlExtractor))
	}

	return "", nil
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
		return nil, err
	}

	me := MetaExtractor{}
	err = m.g.First(&me, "uid = ?", k.Bytes()).Error
	if err != nil {
		return nil, err
	}

	return me.ToDomainModel(m.g)
}

func (m *MSSQLExtractorService) Delete(id string) error {
	return errors.New("Not Implemented")
}
