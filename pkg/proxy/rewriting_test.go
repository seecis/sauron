// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proxy

import (
	"testing"
	"os"
	"github.com/gin-gonic/gin/json"
	"github.com/mkenney/go-chrome/tot/dom"
	"golang.org/x/net/html"
	"io/ioutil"
	"fmt"
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/vanng822/go-premailer/premailer"
)

func TestToHtmlNode(t *testing.T) {
	file, err := os.OpenFile("ebrar.json", os.O_RDONLY, 0)

	if err != nil {
		t.Fatal(err)
	}

	d := dom.GetDocumentResult{}
	err = json.NewDecoder(file).Decode(&d)
	if err != nil {
		t.Fatal(err)
	}

	outfile, err := ioutil.TempFile("","test_")
	if err != nil {
		t.Fatal(err)
	}

	newNode := toHtmlNode(d.Root)
	fmt.Println(outfile.Name())
	err = html.Render(outfile, newNode)
	if err != nil {
		t.Fatal(err)
	}
}


func TestProxy_RequestPage(t *testing.T) {
	proxy := Proxy{
		sessions: nil,
	}


	var buf []byte
	buffer := bytes.NewBuffer(buf)
	proxy.RequestPage("0","https://www.ebrarbilgisayar.com/notebook-pmk189", buffer)

	doc, err := goquery.NewDocumentFromReader(buffer)
	if err != nil {
		t.Fatal(err)
	}


	pm := premailer.NewPremailer(doc, nil)
	str, err := pm.Transform()
	if err != nil {
		t.Fatal(err)
	}

	file, err := ioutil.TempFile("","premail_")
	if err != nil {
		t.Fatal(err)
	}

	_, err = file.WriteString(str)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(file.Name())
}
