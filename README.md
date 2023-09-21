# Colibri
Colibri is an extensible web crawling and scraping framework for Go, used to crawl and extract structured data on the web.

## Installation
```
 $ go get github.com/eduardogxnzalez/colibri
```

## Quick Starts
```go
type (
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
```

## Do
```go
// Do performs an HTTP request according to the rules.
func (c *Colibri) Do(rules *Rules) (resp Response, err error)
```
```go
c := colibri.New()
c.Client = ...    // Required
c.Delay = ...     // Optional
c.RobotsTxt = ... // Optional
c.Parser = ...    // Optional

rules, err := colibri.NewRules(map[string]any{...})
if err != nil {
	panic(err)
}

resp, err := c.Do(rules)
if err != nil {
	panic(err)
}

fmt.Println("URL:", resp.URL())
fmt.Println("Status code:", resp.StatusCode())
fmt.Println("Content-Type", resp.Header().Get("Content-Type"))
```

## Extract
```go
// Extract performs the HTTP request and parses the content of the response following the rules.
// It returns the response of the request, the data extracted with the selectors
// and an error (if any).
func (c *Colibri) Extract(rules *Rules) (resp Response, output map[string]any, err error)
```
```go
var rawRules = []byte(`{...}`) // Raw Rules ~ JSON 

c := colibri.New()
c.Client = ...    // Required
c.Delay = ...     // Optional
c.RobotsTxt = ... // Optional
c.Parser = ...    // Required

var rules colibri.Rules
err := json.Unmarshal(data, &rules)
if err != nil {
	panic(err)
} 

resp, data, err := c.Extract(&rules)
if err != nil {
	panic(err)
}

fmt.Println("URL:", resp.URL())
fmt.Println("Status code:", resp.StatusCode())
fmt.Println("Content-Type", resp.Header().Get("Content-Type"))
fmt.Println("Data:", data)
```

# Raw  Rules ~ JSON
```json
{
	"Method": "string",
	"URL": "string",
	"Proxy": "string",
	"Header": {
		"string": "string",
		"string": ["string", "string", ...]
	},
	"Timeout": "string_or_number",
	"UseCookies": "bool_string_or_number",
	"IgnoreRobotsTxt": "bool_string_or_number",
	"Delay": "string_or_number",
	"Selectors": {...}
}
```

## Selectors
```json
{
	"Selectors": {
		"key_name": "expression"
	}
}
```
```json
{
	"Selectors": {
		"title": "//head/title"
	}
}
```

```json
{
	"Selectors": {
		"key_name":  {
			"Expr": "expression",
			"Type": "expression_type",
			"All": "bool_or_string",
			"Follow": "bool_or_string",
			"Selectors": {...}
		}
	}
}
```
```json
{
	"Selectors": {
		"title":  {
			"Expr": "//head/title",
			"Type": "xpath"
		}
	}
}
```

### Nested selectors
```json
{
	"Selectors": {
		"body":  {
			"Expr": "//body",
			"Type": "xpath",
			"Selectors": {
				"p": "//p"
			}
		}
	}
}
```

### Find all
```json
{
	"Selectors": {
		"a":  {
			"Expr": "//body/a",
			"Type": "xpath",
			"All": true,
		}
	}
}
```

### Follow URLs
```json
{
	"Selectors": {
		"a":  {
			"Expr": "//body/a",
			"Type": "xpath",
			"All": true,
			"Follow": true,
			"Selectors": {
				"title": "//head/title"
			}
		}
	}
}
```

```json
{
	"Selectors": {
		"a":  {
			"Expr": "//body/a",
			"Type": "xpath",
			"All": true,
			"Follow": true,
			"Proxy": "http://proxy-url.com:8080",
			"UseCookies": true,
			"Selectors": {
				"title": "//head/title"
			}
		}
	}
}
```

### Custom fields
```json
{
	"Selectors": {
		"title":  {
			"Expr": "//head/title",
			"Type": "xpath",
			"required": true
		}
	}
}
```

##  Example
```json
{
	"Method": "GET",
	"URL": "https://example.com",
	"Header": {
		"User-Agent": "test/0.0.1",
	},
	"Timeout": 5,
	"Selectors": {
		"a":  {
			"Expr": "//body/a",
			"Type": "xpath",
			"All": true,
			"Follow": true,
			"Selectors": {
				"title": "//head/title"
			}
		}
	}
}
```