package parsers

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/eduardogxnzalez/colibri"
)

type Element interface {
	// Find finds a child element that matches the expression.
	Find(expr, exprType string) (Element, error)

	// FindAll finds all child elements that match the expression.
	FindAll(expr, exprType string) ([]Element, error)

	// Value returns the value of the element.
	Value() any
}

func findSelectors(src *colibri.Rules, resp colibri.Response, selectors []*colibri.Selector, parent Element) (map[string]any, error) {
	if (resp == nil) || (selectors == nil) || (parent == nil) {
		return nil, nil
	}

	var (
		result = make(map[string]any)
		errs   error
	)
	for _, selector := range selectors {
		found, err := findSelector(src, resp, selector, parent)
		if err != nil {
			errs = colibri.AddError(errs, selector.Name, err)
			continue
		}
		result[selector.Name] = found
	}
	return result, errs
}

func followSelector(src *colibri.Rules, resp colibri.Response, selector *colibri.Selector, rawURL ...any) (map[string]any, error) {
	var (
		result = make(map[string]any)
		urls   = make([]*url.URL, 0, len(rawURL))
		errs   error
	)

	for _, rawU := range rawURL {
		u, err := colibri.ToURL(rawU)
		if err != nil {
			errs = colibri.AddError(errs, fmt.Sprintf("%v", rawU), err)
			continue
		}

		if !u.IsAbs() {
			u = resp.URL().ResolveReference(u)
		}
		urls = append(urls, u)
	}

	if errs != nil {
		return nil, errs
	}

	rules := selector.Rules(src)
	for _, u := range urls {
		cRules := rules.Clone()
		cRules.URL = u

		_, found, err := resp.Extract(cRules)
		if err != nil {
			errs = colibri.AddError(errs, u.String(), err)
			continue
		}
		result[u.String()] = found

		colibri.ReleaseRules(cRules)
	}

	colibri.ReleaseRules(rules)
	return result, errs
}

func findAllSelector(src *colibri.Rules, resp colibri.Response, selector *colibri.Selector, parent Element) (any, error) {
	children, err := parent.FindAll(selector.Expr, selector.Type)
	if err != nil {
		return nil, err
	}

	var (
		result []any
		errs   error
	)
	if !selector.Follow && (len(selector.Selectors) > 0) {
		for i, child := range children {
			found, err := findSelectors(src, resp, selector.Selectors, child)
			if err != nil {
				errs = colibri.AddError(errs, selector.Name+"#"+strconv.Itoa(i), err)
				continue
			}
			result = append(result, found)
		}

		return result, errs
	}

	for _, child := range children {
		result = append(result, child.Value())
	}

	if selector.Follow {
		return followSelector(src, resp, selector, result...)
	}
	return result, errs
}

func findSelector(src *colibri.Rules, resp colibri.Response, selector *colibri.Selector, parent Element) (any, error) {
	if (selector == nil) || (parent == nil) {
		return nil, nil
	}

	if selector.All {
		return findAllSelector(src, resp, selector, parent)
	}

	child, err := parent.Find(selector.Expr, selector.Type)
	if err != nil {
		return nil, err
	} else if child == nil {
		return nil, nil
	}

	if selector.Follow {
		return followSelector(src, resp, selector, child.Value())
	}

	if len(selector.Selectors) > 0 {
		return findSelectors(src, resp, selector.Selectors, child)
	}
	return child.Value(), nil
}
