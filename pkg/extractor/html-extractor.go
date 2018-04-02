package extractor

import (
	"bufio"
	"gopkg.in/yaml.v2"
	"io"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"log"
)

type Query struct {
	Selector        string  `yaml:"selector"`
	Name            string  `yaml:"name"`
	ForEachChildren bool    `yaml:"forEachChildren"`
	SubQueries      []Query `yaml:"subQueries"`
	Trim            bool    `yaml:"trim,omitempty"`
}

type HtmlExtractor struct {
	Name    string  `yaml:"name"`
	Queries []Query `yaml:"queries"`
}

func openDocument(reader *bufio.Reader) (Queryable, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}

	return &DocumentWrapper{doc}, nil
}

func (he *HtmlExtractor) Extract(reader *bufio.Reader) (*Field, error) {
	doc, err := openDocument(reader)
	if err != nil {
		return nil, err
	}

	rootQuery := Query{
		Selector:        "",
		Name:            "",
		ForEachChildren: false,
		SubQueries:      he.Queries,
		Trim:            false,
	}

	f, err := executeQuery(doc, rootQuery)
	if err != nil {
		return nil, err
	}

	return &f, nil
}

type SelectionWrapper struct {
	*goquery.Selection
}

type DocumentWrapper struct {
	*goquery.Document
}

type Queryable interface {
	F(string) Queryable
	Text() string
	EachQ(f func(int, Queryable)) Queryable
	ChildrenQ() Queryable
}

func wrapSelection(selection *goquery.Selection) Queryable {
	return &SelectionWrapper{selection}
}

func wrapDocument(document *goquery.Document) Queryable {
	return &DocumentWrapper{document}
}

func (sw *SelectionWrapper) ChildrenQ() Queryable {
	return wrapSelection(sw.Children())
}

func (dw *DocumentWrapper) ChildrenQ() Queryable {
	return wrapSelection(dw.Children())
}

func (sw *SelectionWrapper) EachQ(f func(int, Queryable)) Queryable {
	q := sw.Each(func(i int, selection *goquery.Selection) {
		f(i, &SelectionWrapper{selection})
	})

	return &SelectionWrapper{q}
}

func (dw *DocumentWrapper) EachQ(f func(int, Queryable)) Queryable {
	q := dw.Each(func(i int, selection *goquery.Selection) {
		f(i, &SelectionWrapper{selection})
	})

	return &SelectionWrapper{q}
}

func (sw *SelectionWrapper) F(query string) Queryable {
	return &SelectionWrapper{sw.Find(query)}
}

func (dw *DocumentWrapper) F(query string) Queryable {
	return &SelectionWrapper{dw.Find(query)}
}

func executeQuery(document Queryable, query Query) (Field, error) {
	node := document.F(query.Selector)
	if query.Selector == "" {
		node = document
	}

	var f Field

	if query.ForEachChildren {
		f.label = query.Name

		node.ChildrenQ().EachQ(func(i int, queryable Queryable) {
			var s Field

			if len(query.SubQueries) > 0 {
				for _, v := range query.SubQueries {
					subres, err := executeQuery(queryable, v)
					if err != nil {
						//todo: fix this
						log.Fatal(err)
					}

					s.subfields = append(s.subfields, subres)
				}

				f.subfields = append(f.subfields, s)
				return
			}

			dt := queryable.Text()
			if query.Trim {
				dt = strings.TrimSpace(dt)
			}

			res := Field{label: string(i),
				data: dt,
				subfields: nil,
			}

			f.subfields = append(f.subfields, res)
		})

		return f, nil
	}

	if len(query.SubQueries) == 0 {
		f.label = query.Name
		dt := node.Text()
		if query.Trim {
			dt = strings.TrimSpace(dt)
		}

		f.data = dt
		return f, nil
	}

	for _, v := range query.SubQueries {
		subresult, err := executeQuery(node, v)
		if err != nil {
			return Field{}, err
		}

		f.subfields = append(f.subfields, subresult)
	}

	return f, nil
}

func NewHtmlExtractor(reader io.Reader) Extractor {
	var htmlExtractor HtmlExtractor
	yaml.NewDecoder(reader).Decode(&htmlExtractor)
	return &htmlExtractor
}
