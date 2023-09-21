package parsers

import (
	"strings"

	"github.com/eduardogxnzalez/colibri"

	"github.com/antchfx/jsonquery"
)

// JSONRegexp contains a regular expression that matches the JSON MIME type.
const JSONRegexp = `^application\/(json|x-json|([a-z]+\+json))`

// JSONElement represents a JSON element compatible with XPath expressions.
type JSONElement struct {
	node *jsonquery.Node
}

// ParseJSON parses the content of the response and returns the root element.
func ParseJSON(resp colibri.Response) (*JSONElement, error) {
	root, err := jsonquery.Parse(resp.Body())
	if err != nil {
		return nil, err
	}
	return &JSONElement{root}, nil
}

func (json *JSONElement) Find(expr, exprType string) (Element, error) {
	if (exprType != "") && !strings.EqualFold(exprType, XPathExpr) {
		return nil, ErrExprType
	}

	jsonNode, err := jsonquery.Query(json.node, expr)
	if err != nil {
		return nil, err
	} else if jsonNode == nil {
		return nil, nil
	}

	return &JSONElement{jsonNode}, nil
}

func (json *JSONElement) FindAll(expr, exprType string) ([]Element, error) {
	if (exprType != "") && !strings.EqualFold(exprType, XPathExpr) {
		return nil, ErrExprType
	}

	jsonNodes, err := jsonquery.QueryAll(json.node, expr)
	if err != nil {
		return nil, err
	}

	var elements []Element
	for _, node := range jsonNodes {
		elements = append(elements, &JSONElement{node})
	}
	return elements, nil
}

func (json *JSONElement) Value() any {
	return json.node.Value()
}
