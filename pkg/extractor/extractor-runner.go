// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package extractor

import (
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"context"
	"os"
	"github.com/segmentio/ksuid"
	"log"
	"time"
)

type Runner interface {
	Run(extractor Extractor) (*Field, error)
}

type ChromeDpRunner struct {
	c    *chromedp.CDP
	ctxt context.Context
	url  string
}

func NewChromeDpRunner(c *chromedp.CDP, ctxt context.Context, url string) *ChromeDpRunner {
	return &ChromeDpRunner{
		c:    c,
		ctxt: ctxt,
		url:  url,
	}
}

func (cdpr *ChromeDpRunner) Run(extractor Extractor) (*Field, error) {
	switch extractor.(type) {
	case HtmlExtractor:
		he := extractor.(HtmlExtractor)
		return cdpr.extractHtml(he)
	default:
		return nil, errors.New("Not supported extractor type")
	}
}

func (cdpr *ChromeDpRunner) extractHtml(extractor HtmlExtractor) (*Field, error) {

	var r string

	cdpr.c.Run(cdpr.ctxt, chromedp.Tasks{
		chromedp.Navigate(cdpr.url),
		chromedp.WaitReady("html"),
		chromedp.Sleep(time.Second * 5),
		chromedp.OuterHTML("html", &r, chromedp.BySearch),
	})

	extractionName := "test" + ksuid.New().String() + ".html"
	file, err := os.Create(extractionName)
	if err != nil {
		log.Fatal(err)
	}

	file.WriteString(r)
	file.Close()

	f, err := os.Open(extractionName)
	return extractor.Extract(f)
}
