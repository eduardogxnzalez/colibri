package parsers

import (
	"io"
	"regexp"
	"strings"

	"github.com/eduardogxnzalez/colibri"
)

// TextRegexp contains a regular expression that matches the MIME type plain text.
const TextRegexp = `^text\/plain`

// TextElement represents a Text element compatible with regular expressions.
type TextElement struct {
	data []byte
}

// ParseText parses the content of the response and returns the root element.
func ParseText(resp colibri.Response) (*TextElement, error) {
	b, err := io.ReadAll(resp.Body())
	if err != nil {
		return nil, err
	}
	return &TextElement{b}, nil
}

func (text *TextElement) Find(expr, exprType string) (Element, error) {
	if (exprType != "") && !strings.EqualFold(exprType, RegularExpr) {
		return nil, ErrExprType
	}

	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}

	data := re.Find(text.data)
	return &TextElement{data}, nil
}

func (text *TextElement) FindAll(expr, exprType string) ([]Element, error) {
	if (exprType != "") && !strings.EqualFold(exprType, RegularExpr) {
		return nil, ErrExprType
	}

	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}

	var elements []Element
	for _, data := range re.FindAll(text.data, -1) {
		elements = append(elements, &TextElement{data})
	}
	return elements, nil
}

func (text *TextElement) Value() any {
	return string(text.data)
}
