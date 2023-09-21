package colibri

import (
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	KeyAll = "All"

	KeyExpr = "Expr"

	KeyFollow = "Follow"

	KeyName = "Name"

	KeyType = "Type"
)

var (
	// ErrInvalidSelector is returned when the selector is invalid.
	ErrInvalidSelector = errors.New("invalid selector")

	// ErrInvalidSelectors is returned when the selectors are invalid.
	ErrInvalidSelectors = errors.New("invalid selectors")
)

var selectorPool = sync.Pool{
	New: func() any {
		return &Selector{Fields: make(map[string]any)}
	},
}

type Selector struct {
	// Name selector name.
	Name string

	// Expr stores the selector expression.
	Expr string

	// Type stores the type of the selector expression.
	Type string

	// All specifies whether all elements are to be found.
	All bool

	// Follow specifies whether the URLs found by the selector should be followed.
	Follow bool

	// Selectors nested selectors.
	Selectors []*Selector

	// Fields stores additional data.
	Fields map[string]any
}

func newSelector(name string, rawSelector any, convFunc ConvFunc) (*Selector, error) {
	var (
		selector = selectorPool.Get().(*Selector)
		err      error
	)

	switch selectorValue := rawSelector.(type) {
	case string:
		if selectorValue == "" {
			return nil, nil
		}
		selector.Expr = selectorValue

	case map[string]any:
		delete(selectorValue, KeyName)
		err = processRaw(selectorValue, selector, convFunc)

	default:
		return nil, ErrInvalidSelector
	}

	selector.Name = name
	return selector, err
}

func newSelectors(rawSelectors any, convFunc ConvFunc) ([]*Selector, error) {
	if rawSelectors == nil {
		return nil, nil
	}

	selectorsMap, ok := rawSelectors.(map[string]any)
	if !ok {
		return nil, ErrInvalidSelectors
	}

	var (
		selectors = make([]*Selector, 0, len(selectorsMap))
		errs      error
	)
	for name, value := range selectorsMap {
		if (name == "") || (value == nil) {
			continue
		}

		selector, err := newSelector(name, value, convFunc)
		if err != nil {
			errs = AddError(errs, name, err)
		} else if selector != nil {
			selectors = append(selectors, selector)
		}
	}
	return selectors, errs
}

// Rules returns a Rules with the Selector data.
// Copies the nested selectors from the Selector and
// gets the rest of the data from Fields, if they are
// not in Fields it uses the data from the source Rules.
func (selector *Selector) Rules(src *Rules) *Rules {
	newRules := &Rules{
		Timeout:         src.Timeout,
		UseCookies:      src.UseCookies,
		IgnoreRobotsTxt: src.IgnoreRobotsTxt,
		Delay:           src.Delay,
		Selectors:       CloneSelectors(selector.Selectors),
		Fields:          make(map[string]any),
	}

	if len(selector.Fields) == 0 {
		if src.Proxy != nil {
			newRules.Proxy = src.Proxy.ResolveReference(&url.URL{})
		}
		newRules.Header = src.Header.Clone()

		return newRules
	}

	// METHOD
	if v, ok := selector.Fields[KeyMethod]; ok {
		newRules.Method, _ = v.(string)
	}

	// PROXY
	if v, ok := selector.Fields[KeyProxy]; ok {
		newRules.Proxy, _ = v.(*url.URL)
	} else if src.Proxy != nil {
		newRules.Proxy = src.Proxy.ResolveReference(&url.URL{})
	}

	// HEADER
	if v, ok := selector.Fields[KeyHeader]; ok {
		newRules.Header, _ = v.(http.Header)
	} else {
		newRules.Header = src.Header.Clone()
	}

	// TIMEOUT
	if v, ok := selector.Fields[KeyTimeout]; ok {
		newRules.Timeout, _ = v.(time.Duration)
	}

	// USECOOKIES
	if v, ok := selector.Fields[KeyUseCookies]; ok {
		newRules.UseCookies, _ = v.(bool)
	}

	// IGNOREROBOTSTXT
	if v, ok := selector.Fields[KeyIgnoreRobotsTxt]; ok {
		newRules.IgnoreRobotsTxt, _ = v.(bool)
	}

	// DELAY
	if v, ok := selector.Fields[KeyDelay]; ok {
		newRules.Delay, _ = v.(time.Duration)
	}

	return newRules
}

// Clone returns a copy of the original selector.
// Cloning the Fields field may produce errors, avoid storing pointer.
func (selector *Selector) Clone() *Selector {
	newSelector := &Selector{
		Name:      selector.Name,
		Expr:      selector.Expr,
		Type:      selector.Type,
		All:       selector.All,
		Follow:    selector.Follow,
		Selectors: CloneSelectors(selector.Selectors),
		Fields:    make(map[string]any),
	}

	for key, value := range selector.Fields {
		newSelector.Fields[key] = value
	}
	return newSelector
}

// Clear clears all fields of the selector.
// Selectors are released, see ReleaseSelector.
func (selector *Selector) Clear() {
	selector.Name = ""
	selector.Expr = ""
	selector.Type = ""
	selector.All = false
	selector.Follow = false

	for _, sel := range selector.Selectors {
		ReleaseSelector(sel)
	}
	selector.Selectors = nil

	clear(selector.Fields)
}

// CloneSelectors clones the selectors.
func CloneSelectors(selectors []*Selector) []*Selector {
	var result []*Selector
	for _, sel := range selectors {
		result = append(result, sel.Clone())
	}
	return result
}

// ReleaseRules clears and sends the selector to the selector pool.
func ReleaseSelector(selector *Selector) {
	selector.Clear()
	selectorPool.Put(selector)
}
