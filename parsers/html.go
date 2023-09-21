package parsers

import (
	"strings"

	"github.com/eduardogxnzalez/colibri"

	"github.com/andybalholm/cascadia"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

// HTMLRegexp contains a regular expression that matches the HTML MIME type.
const HTMLRegexp = `^text\/html`

// HTMLElement represents an HTML element compatible with XPath expressions and CSS selectors.
// If the type of expression is not specified, they assume it is an XPath expression.
type HTMLElement struct {
	node *html.Node
}

// ParseHTML parses the content of the response and returns the root element.
func ParseHTML(resp colibri.Response) (*HTMLElement, error) {
	contentType := resp.Header().Get("Content-Type")
	r, err := charset.NewReader(resp.Body(), contentType)
	if err != nil {
		return nil, err
	}

	root, err := htmlquery.Parse(r)
	if err != nil {
		return nil, err
	}
	return &HTMLElement{root}, nil
}

func (html *HTMLElement) Find(expr, exprType string) (Element, error) {
	if exprType == "" {
		exprType = XPathExpr
	}

	switch {
	case strings.EqualFold(exprType, XPathExpr):
		return html.XPathFind(expr)
	case strings.EqualFold(exprType, CSSSelector):
		return html.CSSFind(expr)
	}
	return nil, ErrExprType
}

func (html *HTMLElement) FindAll(expr, exprType string) ([]Element, error) {
	if exprType == "" {
		exprType = XPathExpr
	}

	switch {
	case strings.EqualFold(exprType, XPathExpr):
		return html.XPathFindAll(expr)
	case strings.EqualFold(exprType, CSSSelector):
		return html.CSSFindAll(expr)
	}
	return nil, ErrExprType
}

func (html *HTMLElement) Value() any {
	return htmlquery.InnerText(html.node)
}

func (html *HTMLElement) XPathFind(expr string) (Element, error) {
	htmlNode, err := htmlquery.Query(html.node, expr)
	if err != nil {
		return nil, err
	} else if htmlNode == nil {
		return nil, nil
	}

	return &HTMLElement{htmlNode}, nil
}

func (html *HTMLElement) XPathFindAll(expr string) ([]Element, error) {
	htmlNodes, err := htmlquery.QueryAll(html.node, expr)
	if err != nil {
		return nil, err
	}

	var elements []Element
	for _, node := range htmlNodes {
		elements = append(elements, &HTMLElement{node})
	}
	return elements, nil
}

func (html *HTMLElement) CSSFind(expr string) (Element, error) {
	sel, err := cascadia.Compile(expr)
	if err != nil {
		return nil, err
	}

	htmlNode := cascadia.Query(html.node, sel)
	if htmlNode == nil {
		return nil, nil
	}
	return &HTMLElement{htmlNode}, nil
}

func (html *HTMLElement) CSSFindAll(expr string) ([]Element, error) {
	sel, err := cascadia.Compile(expr)
	if err != nil {
		return nil, err
	}

	var elements []Element
	for _, node := range cascadia.QueryAll(html.node, sel) {
		elements = append(elements, &HTMLElement{node})
	}
	return elements, nil
}
