package parsers

import (
	"strings"

	"github.com/eduardogxnzalez/colibri"

	"github.com/antchfx/xmlquery"
)

// XMLRegexp contains a regular expression that matches the XML MIME type.
const XMLRegexp = `(?i)((application|image|message|model)/((\w|\.|-)+\+?)?|text/)(wb)?xml`

// XMLElement represents an XML element compatible with XPath expressions.
type XMLElement struct {
	node *xmlquery.Node
}

// ParseXML parses the content of the response and returns the root element.
func ParseXML(resp colibri.Response) (*XMLElement, error) {
	root, err := xmlquery.Parse(resp.Body())
	if err != nil {
		return nil, err
	}
	return &XMLElement{root}, nil
}

func (xml *XMLElement) Find(expr, exprType string) (Element, error) {
	if (exprType != "") && !strings.EqualFold(exprType, XPathExpr) {
		return nil, ErrExprType
	}

	xmlNode, err := xmlquery.Query(xml.node, expr)
	if err != nil {
		return nil, err
	} else if xmlNode == nil {
		return nil, nil
	}

	return &XMLElement{xmlNode}, nil
}

func (xml *XMLElement) FindAll(expr, exprType string) ([]Element, error) {
	if (exprType != "") && !strings.EqualFold(exprType, XPathExpr) {
		return nil, ErrExprType
	}

	xmlNodes, err := xmlquery.QueryAll(xml.node, expr)
	if err != nil {
		return nil, err
	}

	var elements []Element
	for _, node := range xmlNodes {
		elements = append(elements, &XMLElement{node})
	}
	return elements, nil
}

func (xml *XMLElement) Value() any {
	return xml.node.InnerText()
}
