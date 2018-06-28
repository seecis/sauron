// Copyright 2018 Legrin, LLC
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proxy

import (
	"net/http"
	"io"
	"github.com/PuerkitoBio/goquery"
	"fmt"
	"strings"
	"github.com/pkg/errors"
	"bytes"
	"golang.org/x/net/html"
	"log"
	"github.com/julienschmidt/httprouter"
	"github.com/segmentio/ksuid"
	"github.com/mkenney/go-chrome/tot"
	"time"
	"github.com/mkenney/go-chrome/tot/dom"
	"golang.org/x/net/html/atom"
	"github.com/mkenney/go-chrome/tot/css"
	"io/ioutil"
	"encoding/json"
	"github.com/vanng822/go-premailer/premailer"
	"net/url"
	"github.com/rs/cors"
)

type Proxy struct {
	sessions *map[string]Session
}

type Session struct {
	baseUrl  string
	proxyUrl string
	pageUrl  string
}

type Handler struct {
}

var browser *chrome.Chrome

func StartProxy() {
	browser = chrome.New(
		&chrome.Flags{
			"addr":                     "localhost",
			"disable-extensions":       nil,
			"disable-gpu":              nil,
			"hide-scrollbars":          nil,
			"no-first-run":             nil,
			"no-sandbox":               nil,
			"port":                     9222,
			"remote-debugging-address": "0.0.0.0",
			"remote-debugging-port":    9222,
		},
		"C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe",
		"",
		"stdout.log",
		"",
	)

	if err := browser.Launch(); nil != err {
		panic(err)
	}

	sessionsMap := make(map[string]Session)
	defaultProxy = Proxy{
		sessions: &sessionsMap,
	}

	router := httprouter.New()
	router.GET("/new", handleCreateSession)
	router.GET("/proxied/:sessId/*req", getResource)
	fmt.Println("Proxy is running at 0.0.0.0:8092")


	http.ListenAndServe("0.0.0.0:8092", cors.AllowAll().Handler(router))
}

var defaultProxy Proxy

func getResource(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	sessionId := params.ByName("sessId")

	sess, ok := (*defaultProxy.sessions)[sessionId]
	if !ok {
		http.Error(w, "No session", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	u, err := url.Parse(sess.pageUrl)
	if err != nil {
		http.Error(w, "Unknown error", http.StatusInternalServerError)
		log.Fatal(err)
	}

	err = defaultProxy.RequestPage(sessionId, u, w)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Unknown error", http.StatusInternalServerError)
		return
	}
	//
	//b, err := ioutil.ReadAll(rd)
	//if err != nil {
	//	log.Fatal(err)
	//	http.Error(w, "Unknown error", http.StatusInternalServerError)
	//	return
	//}
	//
	//w.Write(b)
}

func handleCreateSession(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	url := r.URL.Query().Get("url")
	sessionId, _ := defaultProxy.CreateSession(url, "http://"+r.Host+"/proxied/")
	//todo: add path here.
	http.Redirect(w, r, "/proxied/"+sessionId+"/", http.StatusTemporaryRedirect)
}

func (ph Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

func (p *Proxy) CreateSession(pageUrl string, proxyUrl string) (string, *Session) {

	id := ksuid.New().String()

	u, err := url.Parse(pageUrl)
	if err != nil {
		return "", nil
	}

	baseUrl := u.Scheme + "://" + u.Host

	s := *p.sessions
	sess := Session{
		baseUrl:  baseUrl,
		pageUrl:  pageUrl,
		proxyUrl: proxyUrl,
	}
	s[id] = sess

	p.sessions = &s
	return id, &sess
}

func getNodeType(node *dom.Node) html.NodeType {
	domType := node.NodeType

	if strings.ToLower(node.NodeName) == "script" {
		return html.ElementNode
	}

	switch domType {
	case 1:
		return html.ElementNode
	case 3:
		return html.TextNode
	case 7:
		return html.CommentNode
	case 8:
		return html.CommentNode
	case 9:
		return html.DocumentNode
	case 10:
		return html.CommentNode
	case 11:
		return html.CommentNode

	}

	return 0
}

type Renderer struct {
	rootNode  *dom.Node
	applyFunc func(node *dom.Node)
}

func (*Renderer) render() {

}

func toHtmlNode(node *dom.Node) *html.Node {
	lim := len(node.Attributes) / 2

	var attrs []html.Attribute

	for i := 0; i < lim; i++ {
		startIndex := 2 * i
		attr := html.Attribute{
			Namespace: "",
			Key:       node.Attributes[startIndex],
			Val:       node.Attributes[startIndex+1],
		}

		attrs = append(attrs, attr)
	}

	typ := getNodeType(node)

	var data string

	if typ == html.TextNode {
		data = node.NodeValue
	} else {
		data = strings.ToLower(node.NodeName)
	}

	hnode := html.Node{
		Type:     typ,
		DataAtom: atom.Lookup([]byte(node.NodeName)),
		Data:     data,
		Attr:     attrs,
	}

	if hnode.Data == "script" {
		return nil
	}
	for _, v := range node.Children {
		child := toHtmlNode(v)
		if child == nil {
			continue
		}

		hnode.AppendChild(child)
	}

	return &hnode
}

func removeScript(n *html.Node) {
	// if note is script tag
	if n.Type == html.ElementNode && n.Data == "script" {
		n.Parent.RemoveChild(n)
		return // script tag is gone...
	}
	// traverse DOM
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		removeScript(c)
	}
}

func (p *Proxy) RequestPage(id string, url *url.URL, writer io.Writer) (error) {
	// Open a tab and navigate to the URL you want to screenshot.
	tab, err := browser.NewTab(url.String())
	if nil != err {
		panic(err)
	}

	<-tab.DOM().Enable()
	<-tab.CSS().Enable()
	defer tab.Close()

	var stylesheetHeaders []*css.StyleSheetHeader
	tab.CSS().OnStyleSheetAdded(func(event *css.StyleSheetAddedEvent) {
		header := event.Header
		stylesheetHeaders = append(stylesheetHeaders, header)
		//fmt.Printf("Stylesheet added: id: %s, title: %s, url: %s\n", header.StyleSheetID, header.Title, header.SourceURL)
	})

	time.Sleep(5 * time.Second)
	// Enable Page events for this tab.
	if enableResult := <-tab.Page().Enable(); nil != enableResult.Err {
		panic(enableResult.Err)
	}

	gdr := <-tab.DOM().GetDocument(&dom.GetDocumentParams{
		Depth:  -1,
		Pierce: false,
	})

	d := gdr.Root

	outfile, err := ioutil.TempFile("", "dom_json")
	defer outfile.Close()

	fmt.Println("Writing response from chrome to ", outfile.Name())

	json.NewEncoder(outfile).Encode(gdr)

	h := toHtmlNode(d)

	removeScript(h)
	doc := goquery.NewDocumentFromNode(h)
	doc.Find("link").Each(func(i int, selection *goquery.Selection) {

		for _, node := range selection.Nodes {
			for k, attr := range node.Attr {
				if attr.Key == "href" {
					u := attr.Val
					if strings.Index(u, "/") == 0 {
						base:= url.Scheme + "://" + url.Host
						attr.Val = base + u
					}

					node.Attr[k] = attr
				}
			}
		}
	})

	pm := premailer.NewPremailer(doc, nil)

	final, err := pm.Transform()
	if err != nil {
		return err
	}
	//
	b := bytes.NewBufferString(final)

	_, err = b.WriteTo(writer)

	return err
}

func findHtmlElemByTag(tag string, root *html.Node) *html.Node {
	if root.DataAtom == atom.Lookup([]byte(tag)) {
		return root
	}

	for i := root.FirstChild; i != nil; i = i.NextSibling {
		foundChild := findHtmlElemByTag(tag, i)
		if foundChild != nil {
			return foundChild
		}
	}

	return nil
}

func clearHtml(sessionId string, proxyUrl string, reader io.Reader) (io.Reader, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}

	baseProxyUrl, err := (&url.URL{}).Parse(proxyUrl)

	doc.Find("link[href]").Each(func(i int, selection *goquery.Selection) {
		originalHref, ok := selection.Attr("href")
		if ok {
			if strings.Index(originalHref, "/") == 0 {
				pp, _ := baseProxyUrl.Parse("/" + sessionId + originalHref)
				pps := pp.String()
				selection.SetAttr("href", pps)
			}
		}
	})

	doc.Find("script").Remove()

	return toHtml(doc)
}

func clearCss(sessionId string, reader io.Reader) (io.Reader, error) {
	return nil, errors.New("Not Implemented")
}

func toHtml(s *goquery.Document) (io.Reader, error) {
	var buf bytes.Buffer
	var err error
	if len(s.Nodes) > 0 {
		for c := s.Nodes[0].FirstChild; c != nil; c = c.NextSibling {
			err = html.Render(&buf, c)
			if err != nil {
				return nil, err
			}
		}
	}

	return &buf, err
}
