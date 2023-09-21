# Colibri ~ WebExtractor
WebExtractor are default interfaces for Colibri ready to start crawling or extracting data on the web.

## Quick Starts

### Do
```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/eduardogxnzalez/colibri"
	"github.com/eduardogxnzalez/colibri/webextractor"
)

var rawRules = `{
	"Method": "GET",
	"URL": "https://example.com"
}`

func main() {
	we, err := webextractor.New()
	if err != nil {
		panic(err)
	}

	var rules colibri.Rules
	err = json.Unmarshal([]byte(rawRules), &rules)
	if err != nil {
		panic(err)
	}

	resp, err := we.Do(&rules)
	if err != nil {
		panic(err)
	}

	fmt.Println("URL:", resp.URL())
	fmt.Println("Status code:", resp.StatusCode())
	fmt.Println("Content-Type", resp.Header().Get("Content-Type"))
}
```
```
URL: https://example.com                           
Status code: 200
Content-Type text/html; charset=UTF-8
```

### Extract
```go
package main

import (
	"fmt"

	"github.com/eduardogxnzalez/colibri"
	"github.com/eduardogxnzalez/colibri/webextractor"
)

var rawRules = map[string]any{
	"Method": "GET",
	"URL":    "https://example.com",
	"Selectors": map[string]any{
		"title": "//head/title",
	},
}

func main() {
	we, err := webextractor.New()
	if err != nil {
		panic(err)
	}

	rules, err := colibri.NewRules(rawRules)
	if err != nil {
		panic(err)
	}

	resp, data, err := we.Extract(rules)
	if err != nil {
		panic(err)
	}

	fmt.Println("URL:", resp.URL())
	fmt.Println("Status code:", resp.StatusCode())
	fmt.Println("Content-Type", resp.Header().Get("Content-Type"))
	fmt.Println("Data:", data)
}

```
```
URL: https://example.com
Status code: 200                                  
Content-Type text/html; charset=UTF-8
Data: map[title:Example Domain] 
```