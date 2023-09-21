// parsers are interfaces that Colibri can use to parse the content of the responses.
package parsers

import (
	"errors"
	"regexp"
	"sync"

	"github.com/eduardogxnzalez/colibri"
)

const (
	XPathExpr   = "xpath"
	CSSSelector = "css"
	RegularExpr = "regular"
)

var (
	// ErrNotMatch is returned when the Content-Tyepe does not match the Paser.
	ErrNotMatch = errors.New("Content-Type does not match")

	// ErrExprType is returned when the type of expression is not compatible with the element.
	ErrExprType = errors.New("ExprType not compatible with Element")
)

// ParserFunc parses the content of the response and returns the root element.
type ParserFunc func(colibri.Response) (Element, error)

// Parsers stores ParserFunc used to parse the content of the responses.
// ParserFunc are stored with a regular expression that functions as a key.
// When a regular expression matches the Content-Type of the response, the content of the response is parsed with the ParserFunc corresponding to the regular expression.
type Parsers struct {
	rw    sync.RWMutex
	funcs map[string]struct {
		re         *regexp.Regexp
		parserFunc ParserFunc
	}
}

// New returns a new Parsers with ParserFunc to parse HTML, XHML, JSON and Plain Text.
// See the colibri.Parser interface.
func New() (*Parsers, error) {
	parsers := &Parsers{
		funcs: make(map[string]struct {
			re         *regexp.Regexp
			parserFunc ParserFunc
		}),
	}

	var errs error
	errs = errors.Join(errs, Set(parsers, HTMLRegexp, ParseHTML))
	errs = errors.Join(errs, Set(parsers, JSONRegexp, ParseJSON))
	errs = errors.Join(errs, Set(parsers, TextRegexp, ParseText))
	errs = errors.Join(errs, Set(parsers, XMLRegexp, ParseXML))

	return parsers, errs
}

// Match returns true if the Content-Type is compatible with the Parser.
func (parsers *Parsers) Match(contentType string) bool {
	parsers.rw.Lock()
	defer parsers.rw.Unlock()

	for _, p := range parsers.funcs {
		if p.re.MatchString(contentType) {
			return true
		}
	}
	return false
}

// Parse parses the response based on the rules.
func (parsers *Parsers) Parse(rules *colibri.Rules, resp colibri.Response) (map[string]any, error) {
	if (rules == nil) || (resp == nil) {
		return nil, nil
	}

	contentType := resp.Header().Get("Content-Type")

	var parserFunc ParserFunc
	parsers.rw.Lock()
	for _, p := range parsers.funcs {
		if p.re.MatchString(contentType) {
			parserFunc = p.parserFunc
			break
		}
	}
	parsers.rw.Unlock()

	if parserFunc == nil {
		return nil, ErrNotMatch
	}

	parent, err := parserFunc(resp)
	if err != nil {
		return nil, err
	}

	return findSelectors(rules, resp, rules.Selectors, parent)
}

// Clear deletes all stored ParserFunc.
func (parsers *Parsers) Clear() {
	parsers.rw.Lock()
	clear(parsers.funcs)
	parsers.rw.Unlock()
}

// Set adds to parsers the regular expression and the corresponding ParserFunc.
func Set[T Element](parsers *Parsers, expr string, parserFunc func(colibri.Response) (T, error)) error {
	if parsers == nil || expr == "" || parserFunc == nil {
		return nil
	}

	regular, err := regexp.Compile(expr)
	if err != nil {
		return err
	}

	parsers.rw.Lock()
	parsers.funcs[expr] = struct {
		re         *regexp.Regexp
		parserFunc ParserFunc
	}{
		re: regular,
		parserFunc: func(resp colibri.Response) (Element, error) {
			return parserFunc(resp)
		},
	}
	parsers.rw.Unlock()
	return nil
}
