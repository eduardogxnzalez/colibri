// Colibri is an extensible web crawling and scraping framework for Go,
// used to crawl and extract structured data on the web.
package colibri

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultUserAgent is the default User-Agent used for requests.
const DefaultUserAgent = "colibri/0.1"

var (
	// ErrClientIsNil returned when Client is nil.
	ErrClientIsNil = errors.New("Client is nil")

	// ErrParserIsNil returned when Parser is nil.
	ErrParserIsNil = errors.New("Parser is nil")

	// ErrRulesIsNil returned when rules are nil.
	ErrRulesIsNil = errors.New("Rules is nil")
)

type (
	// Response represents an HTTP response.
	Response interface {
		// URL returns the URI of the request used to obtain the response.
		URL() *url.URL

		// StatusCode returns the status code.
		StatusCode() int

		// Header returns the HTTP header of the response.
		Header() http.Header

		// Body returns the response body.
		Body() io.ReadCloser

		// Do Colibri Do method wrapper.
		// Wraps the Colibri used to obtain the HTTP response.
		Do(rules *Rules) (Response, error)

		// Extract Colibri Extract method wrapper.
		// Wraps the Colibri used to obtain the HTTP response.
		Extract(rules *Rules) (Response, map[string]any, error)
	}

	// HTTPClient represents an HTTP client.
	HTTPClient interface {
		// Do makes HTTP requests.
		Do(c *Colibri, rules *Rules) (Response, error)

		// Clear cleans the fields of the structure.
		Clear()
	}

	// Delay manages the delay between each HTTP request.
	Delay interface {
		// Wait waits for the previous HTTP request to the same URL and stores
		// the timestamp, then starts the calculated delay with the timestamp
		// and the specified duration of the delay.
		Wait(u *url.URL, duration time.Duration)

		// Done warns that an HTTP request has been made to the URL.
		Done(u *url.URL)

		// Stamp records the time at which the HTTP request to the URL was made.
		Stamp(u *url.URL)

		// Clear cleans the fields of the structure.
		Clear()
	}

	// RobotsTxt represents a robots.txt parser.
	RobotsTxt interface {
		// IsAllowed verifies that the User-Agent can access the URL.
		IsAllowed(c *Colibri, rules *Rules) error

		// Clear cleans the fields of the structure.
		Clear()
	}

	// Parser represents a parser of the response content.
	Parser interface {
		// Match returns true if the Content-Type is compatible with the Parser.
		Match(contentType string) bool

		// Parse parses the response based on the rules.
		Parse(rules *Rules, resp Response) (map[string]any, error)

		// Clear cleans the fields of the structure.
		Clear()
	}
)

// Colibri performs HTTP requests and parses
// the content of the response based on rules.
type Colibri struct {
	Client    HTTPClient
	Delay     Delay
	RobotsTxt RobotsTxt
	Parser    Parser
}

// New returns a new empty Colibri structure.
func New() *Colibri {
	return &Colibri{}
}

// Do performs an HTTP request according to the rules.
func (c *Colibri) Do(rules *Rules) (resp Response, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	if c.Client == nil {
		return nil, ErrClientIsNil
	}

	if rules == nil {
		return nil, ErrRulesIsNil
	}

	if rules.Header == nil {
		rules.Header = http.Header{}
	}

	if strings.TrimSpace(rules.Header.Get("User-Agent")) == "" {
		rules.Header.Set("User-Agent", DefaultUserAgent)
	}

	if (c.RobotsTxt != nil) && !rules.IgnoreRobotsTxt {
		err := c.RobotsTxt.IsAllowed(c, rules)
		if err != nil {
			return nil, err
		}
	}

	if (c.Delay != nil) && (rules.Delay > 0) {
		c.Delay.Wait(rules.URL, rules.Delay)
		defer c.Delay.Done(rules.URL)
	}

	resp, err = c.Client.Do(c, rules)

	if (c.Delay != nil) && (resp != nil) {
		c.Delay.Stamp(resp.URL())
	}
	return resp, err
}

// Extract performs the HTTP request and parses the content of the response following the rules.
// It returns the response of the request, the data extracted with the selectors
// and an error (if any).
func (c *Colibri) Extract(rules *Rules) (resp Response, output map[string]any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	if c.Parser == nil {
		return nil, nil, ErrParserIsNil
	}

	resp, err = c.Do(rules)
	if err != nil {
		return nil, nil, err
	}

	if len(rules.Selectors) > 0 {
		output, err = c.Parser.Parse(rules, resp)
	}
	return resp, output, err
}

// Clear cleans the fields of the structure.
func (c *Colibri) Clear() {
	if c.Client != nil {
		c.Client.Clear()
	}

	if c.Delay != nil {
		c.Delay.Clear()
	}

	if c.RobotsTxt != nil {
		c.RobotsTxt.Clear()
	}

	if c.Parser != nil {
		c.Parser.Clear()
	}
}
