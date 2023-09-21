package colibri

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"time"
)

const (
	KeyDelay = "Delay"

	KeyFields = "Fields"

	KeyHeader = "Header"

	KeyIgnoreRobotsTxt = "IgnoreRobotsTxt"

	KeyMethod = "Method"

	KeyProxy = "Proxy"

	KeySelectors = "Selectors"

	KeyTimeout = "Timeout"

	KeyUseCookies = "UseCookies"

	KeyURL = "URL"
)

// ErrNotAssignable is returned when the value of RawRules cannot be assigned to the structure field.
var ErrNotAssignable = errors.New("value is not assignable to field")

var rulesPool = sync.Pool{
	New: func() any {
		return &Rules{Fields: make(map[string]any)}
	},
}

// RawRules represents the raw rules.
type RawRules map[string]any

type Rules struct {
	// Method specifies the HTTP method (GET, POST, PUT, ...).
	Method string

	// URL specifies the requested URI.
	URL *url.URL

	// Proxy specifies the proxy URI.
	Proxy *url.URL

	// Header contains the HTTP header.
	Header http.Header

	// Timeout specifies the time limit for the HTTP request.
	Timeout time.Duration

	// UseCookies specifies whether the client should send and store Cookies.
	UseCookies bool

	// IgnoreRobotsTxt specifies whether robots.txt should be ignored.
	IgnoreRobotsTxt bool

	// Delay specifies the delay time between requests.
	Delay time.Duration

	// Selectors
	Selectors []*Selector

	// Fields stores additional data.
	Fields map[string]any
}

// NewRules returns the rules processed using DefaultConvFunc.
func NewRules(rawRules RawRules) (*Rules, error) {
	return NewRulesWithConvFunc(rawRules, DefaultConvFunc)
}

// NewRulesWithConvFunc returns the processed rules.
func NewRulesWithConvFunc(rawRules RawRules, convFunc ConvFunc) (*Rules, error) {
	newRules := rulesPool.Get().(*Rules)
	err := processRaw(rawRules, newRules, convFunc)
	return newRules, err
}

// Clone returns a copy of the original rules.
// Cloning the Fields field may produce errors, avoid storing pointer.
func (rules *Rules) Clone() *Rules {
	newRules := &Rules{
		Method:          rules.Method,
		Header:          rules.Header.Clone(),
		Timeout:         rules.Timeout,
		UseCookies:      rules.UseCookies,
		IgnoreRobotsTxt: rules.IgnoreRobotsTxt,
		Delay:           rules.Delay,
		Selectors:       CloneSelectors(rules.Selectors),
		Fields:          make(map[string]any),
	}

	if rules.URL != nil {
		newRules.URL = rules.URL.ResolveReference(&url.URL{})
	}

	if rules.Proxy != nil {
		newRules.Proxy = rules.Proxy.ResolveReference(&url.URL{})
	}

	for key, value := range rules.Fields {
		newRules.Fields[key] = value
	}
	return newRules
}

// Clear clears all fields of the rules.
// Selectors are released, see ReleaseSelector.
func (rules *Rules) Clear() {
	rules.Method = ""
	rules.URL = nil
	rules.Proxy = nil
	rules.Header = nil
	rules.Timeout = 0

	rules.UseCookies = false
	rules.IgnoreRobotsTxt = false
	rules.Delay = 0

	for _, sel := range rules.Selectors {
		ReleaseSelector(sel)
	}
	rules.Selectors = nil

	clear(rules.Fields)
}

func (rules *Rules) UnmarshalJSON(b []byte) error {
	rawRules := make(map[string]any)
	if err := json.Unmarshal(b, &rawRules); err != nil {
		return err
	}

	newRules, err := NewRules(rawRules)
	if err != nil {
		return err
	}

	*rules = *newRules
	return nil
}

func processRaw[T Rules | Selector](raw map[string]any, output *T, convFunc ConvFunc) error {
	if raw == nil {
		return nil
	}

	var (
		rOutput = reflect.ValueOf(output)
		errs    error
	)
	for key, value := range raw {
		if convFunc != nil {
			var err error

			value, err = convFunc(key, value)
			if err != nil {
				errs = AddError(errs, key, err)
				continue
			}
		}
		rValue := reflect.ValueOf(value)

		field := rOutput.Elem().FieldByName(key)
		if field.IsValid() && field.CanSet() {
			if rValue.Type().AssignableTo(field.Type()) {
				field.Set(rValue)
				continue
			}

			errs = AddError(errs, key, ErrNotAssignable)
			continue
		}

		// Fields
		field = rOutput.Elem().FieldByName(KeyFields)
		if field.IsValid() && field.CanSet() && (field.Kind() == reflect.Map) {
			if field.IsNil() {
				field.Set(reflect.MakeMap(field.Type()))
			}

			field.SetMapIndex(reflect.ValueOf(key), rValue)
		}
	}
	return errs
}

// ReleaseRules clears and sends the rules to the rules pool.
func ReleaseRules(rules *Rules) {
	rules.Clear()
	rulesPool.Put(rules)
}
