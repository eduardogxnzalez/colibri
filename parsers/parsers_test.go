package parsers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/eduardogxnzalez/colibri"
)

func TestParsers(t *testing.T) {
	parsers, err := New()
	if err != nil {
		t.Fatal(err)
	}

	c := colibri.New()
	c.Client = &testClient{}
	c.Parser = parsers

	var emptySlice []any

	tests := []struct {
		Name   string
		Rules  *colibri.Rules
		Output map[string]any
		ErrMap map[string]any
	}{
		{
			"HTML",
			&colibri.Rules{
				Selectors: []*colibri.Selector{
					{Name: "title", Expr: "title", Type: "css"},
					{Name: "p", Expr: "p", Type: "css"}, // Does not exist

					{Name: "a", Expr: "//a/text()", All: true},
					{
						Name: "a-url",
						Expr: "a",
						Type: "css",
						All:  true,
						Selectors: []*colibri.Selector{
							{Name: "url", Expr: "/@href"},
							{Name: "id", Expr: "/@id", Type: "xpath"}, // Does not exist
						},
					},
					{
						Name:   "a-follow",
						Expr:   "//a/@href",
						All:    true,
						Follow: true,
						Selectors: []*colibri.Selector{
							{Name: "title", Expr: "title", Type: "css"},
						},
						Fields: map[string]any{
							"Header": http.Header{"Accept": []string{"text/html"}},
							"Fields": "ignore", // ignore
						},
					},
					{Name: "span", Expr: "//span", Type: "xpath", All: true}, // Does not exist
					{Name: "divs", Expr: "div", Type: "css", All: true},      // Does not exist
				},
				Fields: map[string]any{
					"Content-Type": "text/html",
					"Body":         htmlBody,
				},
			},
			map[string]any{
				"title": "My test page",
				"p":     nil,
				"a":     []any{"Link 1", "Link 2", "Link 3"},
				"a-url": []any{
					map[string]any{"id": nil, "url": "https://page.test/html/1"},
					map[string]any{"id": nil, "url": "https://page.test/html/2"},
					map[string]any{"id": nil, "url": "https://page.test/html/3"},
				},
				"a-follow": map[string]any{
					"https://page.test/html/1": map[string]any{
						"title": "My test page",
					},
					"https://page.test/html/2": map[string]any{
						"title": "My test page",
					},
					"https://page.test/html/3": map[string]any{
						"title": "My test page",
					},
				},
				"span": emptySlice,
				"divs": emptySlice,
			},
			nil,
		},
		{
			"JSON",
			&colibri.Rules{
				Selectors: []*colibri.Selector{
					{Name: "name", Expr: "//name"},
					{
						Name: "contact",
						Expr: "//contact",
						Type: "xpath",
						Selectors: []*colibri.Selector{
							{Name: "phone", Expr: "//phone", Type: "xpath"}, // Does not exist
							{
								Name:   "web",
								Expr:   "//web",
								Type:   "xpath",
								Follow: true,
								Selectors: []*colibri.Selector{
									{Name: "url", Expr: "//URL", Type: "xpath"},
								},
								Fields: map[string]any{
									"Header": http.Header{"Accept": []string{"application/json"}},
									"URL":    123, // ignore
								},
							},
						},
					},

					{Name: "hobbies", Expr: "//hobbies/*", All: true},
					{Name: "jobs", Expr: "//jobs/*", Type: "xpath", All: true}, // Does not exist
				},
				Fields: map[string]any{
					"Content-Type": "application/json",
					"Body":         jsonBody,
				},
			},
			map[string]any{
				"name": "Go Gopher",
				"contact": map[string]any{
					"phone": nil,
					"web": map[string]any{
						"https://go.dev/blog/gopher": map[string]any{
							"url": "https://go.dev/blog/gopher",
						},
					},
				},

				"hobbies": []any{"coding", "backend"},
				"jobs":    emptySlice,
			},
			nil,
		},
		{
			"Text",
			&colibri.Rules{
				Selectors: []*colibri.Selector{
					{
						Name: "21",
						Expr: `^.{0,21}`,
						Type: "regular",
						Selectors: []*colibri.Selector{
							{Name: "first-a", Expr: `\ba\w+`, Type: "regular"}, // Does not exist
							{Name: "first-B", Expr: `\bB\w+`},
						},
					},
					{
						Name:   "source",
						Expr:   `\bhttps?://\S+/source\b`,
						Type:   "regular",
						Follow: true,
						Selectors: []*colibri.Selector{
							{Name: "link", Expr: `\bhttps?://\S+\b`, Type: "regular"},
						},
						Fields: map[string]any{
							"Header": http.Header{"Accept": []string{"text/plain"}},
						},
					},

					{Name: "go", Expr: `\bGo\b`, Type: "regular", All: true},
					{Name: "emails", Expr: `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`, Type: "regular", All: true}, // Does not exist
					{
						Name: "and",
						Expr: `\band\s+\w+`,
						All:  true,
						Selectors: []*colibri.Selector{
							{Name: "sig", Expr: `\s+\w+`, Type: "regular"},
						},
					},
					{
						Name:   "URLs",
						Expr:   `\bhttps?://\S+\b`,
						Type:   "regular",
						All:    true,
						Follow: true,
						Selectors: []*colibri.Selector{
							{Name: "link", Expr: `.*`, Type: "regular"},
						},
						Fields: map[string]any{
							"Header":    http.Header{"Accept": []string{"text/plain"}},
							"Selectors": nil, // ignore
						},
					},
				},
				Fields: map[string]any{
					"Content-Type": "text/plain",
					"Body":         textBody,
				},
			},
			map[string]any{
				"21": map[string]any{
					"first-a": "",
					"first-B": "Binary",
				},
				"source": map[string]any{
					"https://go.dev/doc/install/source": map[string]any{
						"link": "https://go.dev/doc/install/source",
					},
				},

				"go":     []any{"Go", "Go", "Go"},
				"emails": emptySlice,
				"and": []any{
					map[string]any{"sig": " architecture"},
					map[string]any{"sig": " proposals"},
				},
				"URLs": map[string]any{
					"https://go.dev/dl":                 map[string]any{"link": "URL: https://go.dev/dl"},
					"https://go.dev/doc/install":        map[string]any{"link": "URL: https://go.dev/doc/install"},
					"https://go.dev/doc/install/source": map[string]any{"link": "URL: https://go.dev/doc/install/source"},
					"https://go.dev/doc/contribute":     map[string]any{"link": "URL: https://go.dev/doc/contribute"},
					"https://go.dev/wiki/Questions":     map[string]any{"link": "URL: https://go.dev/wiki/Questions"},
				},
			},
			nil,
		},
		{
			"XML",
			&colibri.Rules{
				Selectors: []*colibri.Selector{
					{
						Name: "channel",
						Expr: "//channel",
						Type: "xpath",
						Selectors: []*colibri.Selector{
							{Name: "title", Expr: "//title"},
							{Name: "language", Expr: "//language"}, // Does not exist
						},
					},
					{
						Name:   "link",
						Expr:   "//channel/link",
						Type:   "xpath",
						Follow: true,
						Selectors: []*colibri.Selector{
							{Name: "title", Expr: "//title", Type: "xpath"},
						},
						Fields: map[string]any{
							"Header": http.Header{"Accept": []string{"application/xml"}},
						},
					},

					{Name: "category", Expr: "//category", All: true},
					{
						Name: "items",
						Expr: "//channel/item",
						Type: "xpath",
						All:  true,
						Selectors: []*colibri.Selector{
							{Name: "title", Expr: "//title", Type: "xpath"},
							{Name: "source", Expr: "//source/@url", Type: "xpath", All: true}, // Does not exist
						},
					},
					{
						Name:   "itemLinks",
						Expr:   "//item/link",
						Type:   "xpath",
						All:    true,
						Follow: true,
						Selectors: []*colibri.Selector{
							{Name: "title", Expr: "//title", Type: "xpath"},
						},
						Fields: map[string]any{
							"Header": http.Header{"Accept": []string{"application/xml"}},
							"URL":    "", // ignore
						},
					},
				},
				Fields: map[string]any{
					"Content-Type": "application/xml",
					"Body":         xmlBody,
				},
			},
			map[string]any{
				"channel": map[string]any{
					"title":    "Test RSS",
					"language": nil,
				},
				"link": map[string]any{
					"https://www.test.rss": map[string]any{
						"title": "Test RSS",
					},
				},

				"category": []any{"testing", "example"},
				"items": []any{
					map[string]any{
						"title":  "Item 2",
						"source": emptySlice,
					},
					map[string]any{
						"title":  "Item 1",
						"source": emptySlice,
					},
				},
				"itemLinks": map[string]any{
					"https://www.test.rss/item1": map[string]any{"title": "Test RSS"},
					"https://www.test.rss/item2": map[string]any{"title": "Test RSS"},
				},
			},
			nil,
		},

		// errors
		{
			"HTMLErr",
			&colibri.Rules{
				Selectors: []*colibri.Selector{
					{Name: "Title", Expr: "]title(", Type: "css"},       // invalid css selector
					{Name: "First", Expr: "//a[text()=", Type: "xpath"}, // invalid XPath
					{Name: "Img", Expr: "img", Type: "error"},           // ErrExprType

					{Name: "a", Expr: "//a[@href==]", Type: "xpath", All: true}, // invalid XPath
					{Name: "Span", Expr: "]@span", Type: "css", All: true},      // invalid css selector
					{Name: "Divs", Expr: "div", Type: "error", All: true},       // ErrExprType
				},
				Fields: map[string]any{
					"Content-Type": "text/html",
					"Body":         htmlBody,
				},
			},
			nil,
			map[string]any{
				"Title": "expected identifier, found ] instead",
				"First": "expression must evaluate to a node-set",
				"Img":   ErrExprType.Error(),

				"a":    "expression must evaluate to a node-set",
				"Span": "expected identifier, found ] instead",
				"Divs": ErrExprType.Error(),
			},
		},
		{
			"JSONErr",
			&colibri.Rules{
				Selectors: []*colibri.Selector{
					{Name: "Female", Expr: ")//female)", Type: "xpath"}, // invalid XPath
					{Name: "City", Expr: "//city", Type: "error"},       // ErrExprType

					{Name: "Hobbies", Expr: "//hobbies[/*", Type: "xpath", All: true}, // invalid XPath
					{Name: "Jobs", Expr: "//job/*", Type: "error", All: true},         // ErrExprType
				},
				Fields: map[string]any{
					"Content-Type": "application/json",
					"Body":         jsonBody,
				},
			},
			nil,
			map[string]any{
				"Female":  "expression must evaluate to a node-set",
				"City":    ErrExprType.Error(),
				"Hobbies": "//hobbies[/* has an invalid token",
				"Jobs":    ErrExprType.Error(),
			},
		},
		{
			"TextErr",
			&colibri.Rules{
				Selectors: []*colibri.Selector{
					{Name: "Go", Expr: `)\bGo\]`, Type: "regular"},                   // invalid regular expression
					{Name: "Source", Expr: `\bhttps?://\S+/source\b`, Type: "error"}, // ErrExprType

					{Name: "URLs", Expr: `\Khttps?://\S+\K`, Type: "regular", All: true},                // invalid regular expression
					{Name: "Emails", Expr: `[\w\.-]+@[\w\.-]+\.[a-zA-Z]{2,}`, Type: "error", All: true}, // ErrExprType
				},
				Fields: map[string]any{
					"Content-Type": "text/plain",
					"Body":         textBody,
				},
			},
			nil,
			map[string]any{
				"Go":     "error parsing regexp: unexpected ): `)\\bGo\\]`",
				"Source": ErrExprType.Error(),
				"URLs":   "error parsing regexp: invalid escape sequence: `\\K`",
				"Emails": ErrExprType.Error(),
			},
		},
		{
			"XMLErr",
			&colibri.Rules{
				Selectors: []*colibri.Selector{
					{Name: "title", Expr: "]//channel[/title", Type: "xpath"}, // invalid XPath
					{Name: "link", Expr: "//link", Type: "error"},             // ErrExprType

					{Name: "items", Expr: "()//channel/item", Type: "xpath", All: true}, // invalid XPath
					{Name: "a", Expr: "//a", Type: "error", All: true},                  // ErrExprType
				},
				Fields: map[string]any{
					"Content-Type": "application/xml",
					"Body":         xmlBody,
				},
			},
			nil,
			map[string]any{
				"title": "expression must evaluate to a node-set",
				"link":  ErrExprType.Error(),

				"items": "expression must evaluate to a node-set",
				"a":     ErrExprType.Error(),
			},
		},

		{
			"JSON_URLErr",
			&colibri.Rules{
				Selectors: []*colibri.Selector{
					{
						Name:   "since",
						Expr:   "//since",
						Type:   "xpath",
						Follow: true,
					},
				},
				Fields: map[string]any{
					"Content-Type": "application/json",
					"Body":         jsonBody,
				},
			},
			nil,
			map[string]any{
				"since": map[string]any{
					"2011": colibri.ErrMustBeString.Error(),
				},
			},
		},
		{
			"XML_URLErr",
			&colibri.Rules{
				Selectors: []*colibri.Selector{
					{
						Name:   "itemLinks",
						Expr:   "//item/link",
						Type:   "xpath",
						All:    true,
						Follow: true,
						Selectors: []*colibri.Selector{
							{Name: "title", Expr: "//title", Type: "xpath"},
						},
						Fields: map[string]any{
							"Header": http.Header{"Accept": []string{"NOTFOUND"}},
						},
					},
				},
				Fields: map[string]any{
					"Content-Type": "application/xml",
					"Body":         xmlBody,
				},
			},
			nil,
			map[string]any{
				"itemLinks": map[string]any{
					"https://www.test.rss/item1": "Not Found",
					"https://www.test.rss/item2": "Not Found",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			resp := newTestResponse(c, tt.Rules)
			output, err := parsers.Parse(tt.Rules, resp)
			if (err != nil) && (tt.ErrMap != nil) {
				wantErr, _ := json.Marshal(tt.ErrMap)
				jsonErrs, _ := json.Marshal(err)

				if !reflect.DeepEqual(wantErr, jsonErrs) {
					t.Fatal(err)
				}
				return

			} else if (err == nil) && (tt.ErrMap == nil) {
				if !reflect.DeepEqual(output, tt.Output) {
					t.Fatal("not equal")
				}
				return
			}

			t.Fatal(err)
		})
	}

	// Response Nil
	t.Run("ResponseNil", func(t *testing.T) {
		output, err := parsers.Parse(&colibri.Rules{}, nil)
		if output != nil {
			t.Fatal("must be nil")
		} else if err != nil {
			t.Fatal(err)
		}
	})

	// ErrNotMatch
	t.Run("ErrNotMatch", func(t *testing.T) {
		rules := &colibri.Rules{
			Fields: map[string]any{
				"Content-Type": "apk",
				"Body":         "",
			},
		}
		resp := newTestResponse(c, rules)
		_, err := parsers.Parse(&colibri.Rules{}, resp)
		if !errors.Is(err, ErrNotMatch) {
			t.Fatal(err)
		}
	})
}

func TestParsersClear(t *testing.T) {
	parsers, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if !parsers.Match("text/plain") {
		t.Fatal(ErrNotMatch)
	}

	parsers.Clear()

	if parsers.Match("text/plain") {
		t.Fatal("must not match")
	}

	if len(parsers.funcs) > 0 {
		t.Fatal("uncleaned map")
	}

	// Set
	t.Run("SetNilFunc", func(t *testing.T) {
		var parserFunc func(colibri.Response) (Element, error)
		err := Set(parsers, ".*", parserFunc)
		if err != nil {
			t.Fatal(err)
		} else if len(parsers.funcs) > 0 {
			t.Fatal("stored nil function")
		}
	})

	t.Run("SetInvalidExpr", func(t *testing.T) {
		err := Set(parsers, `[abc`, ParseXML)
		if err == nil {
			t.Fatal("invalid regular expression stored")
		}
	})
}

const (
	htmlBody = `<!doctype html>
	<html>
  	<head>
    	<title>My test page</title>
    </head>
    <body>
  		<a href="https://page.test/html/1">Link 1</a>
  		<a href="https://page.test/html/2">Link 2</a>
  		<a href="https://page.test/html/3">Link 3</a>
  	</body>
	</html>`

	jsonBody = `{
		"name": "Go Gopher",
		"since": 2011,
		"contact": {
			"web": "https://go.dev/blog/gopher"
		},
		"hobbies": [
			"coding",
			"backend"
		]
	}`

	textBody = `	Binary Distributions
		Official binary distributions are available at https://go.dev/dl/.
		After downloading a binary release, visit https://go.dev/doc/install for installation instructions.

	Install From Source
		If a binary distribution is not available for your combination of operating system and architecture,
		visit https://go.dev/doc/install/source for source installation instructions.

	Contributing
		Go is the work of thousands of contributors. We appreciate your help!
		To contribute, please read the contribution guidelines at https://go.dev/doc/contribute.
		Note that the Go project uses the issue tracker for bug reports and proposals only.
		See https://go.dev/wiki/Questions for a list of places to ask questions about the Go language.`

	xmlBody = `<?xml version="1.0" encoding="UTF-8" ?>
	<rss version="2.0">
		<channel>
  		<title>Test RSS</title>
  		<link>https://www.test.rss</link>
  		<category>testing</category>
		<category>example</category> 
  		
  		<item>
    		<title>Item 2</title>
    		<link>https://www.test.rss/item2</link>
  		</item>
  		<item>
    		<title>Item 1</title>
    		<link>https://www.test.rss/item1</link>
  		</item>
  	</channel>
	</rss>`
)

type testResp struct {
	u      *url.URL
	header http.Header
	body   io.ReadCloser
	c      *colibri.Colibri
}

func newTestResponse(c *colibri.Colibri, rules *colibri.Rules) *testResp {
	contentType := rules.Fields["Content-Type"].(string)
	body := rules.Fields["Body"].(string)

	resp := &testResp{u: rules.URL, header: http.Header{}, c: c}

	resp.header.Set("Content-Type", contentType)
	resp.body = io.NopCloser(strings.NewReader(body))
	return resp
}

func (r *testResp) URL() *url.URL                                     { return r.u }
func (r *testResp) StatusCode() int                                   { return 200 }
func (r *testResp) Header() http.Header                               { return r.header }
func (r *testResp) Body() io.ReadCloser                               { return r.body }
func (r *testResp) Do(rules *colibri.Rules) (colibri.Response, error) { return r.c.Do(rules) }
func (r *testResp) Extract(rules *colibri.Rules) (colibri.Response, map[string]any, error) {
	return r.c.Extract(rules)
}

type testClient struct{}

func (client *testClient) Do(c *colibri.Colibri, rules *colibri.Rules) (colibri.Response, error) {
	var (
		accept = rules.Header.Get("Accept")
		body   string
	)

	switch {
	case regexp.MustCompile(HTMLRegexp).MatchString(accept):
		body = htmlBody

	case regexp.MustCompile(JSONRegexp).MatchString(accept):
		body = "{\"URL\": \"" + rules.URL.String() + "\"}"

	case regexp.MustCompile(TextRegexp).MatchString(accept):
		body = "URL: " + rules.URL.String()

	case regexp.MustCompile(XMLRegexp).MatchString(accept):
		body = xmlBody

	default:
		return nil, errors.New("Not Found")

	}

	rules.Fields["Content-Type"] = accept
	rules.Fields["Body"] = body
	return newTestResponse(c, rules), nil
}

func (client *testClient) Clear() {}
