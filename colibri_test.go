package colibri

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestColibriDo(t *testing.T) {
	var (
		c      = New()
		client = &testClient{}
		delay  = &testDelay{}
		robots = &testRobots{}

		testErr = errors.New("Test Error")
	)

	tests := []struct {
		Name   string
		Rules  *Rules
		Client HTTPClient
		Delay  Delay
		Robots RobotsTxt

		DelayWaitUsed  bool
		DelayStampUsed bool
		RobotsUsed     bool
		Err            error
	}{
		{"OK", &Rules{Delay: time.Second}, client, delay, robots, true, true, true, nil},
		{"ClientIsNil", &Rules{}, nil /*Client*/, delay, robots, false, false, false, ErrClientIsNil},
		{"RulesIsNil", nil /*Rules*/, client, delay, robots, false, false, false, ErrRulesIsNil},

		{"NoDelay", &Rules{}, client, nil /*Delay*/, robots, false, false, true, nil},
		{"NoDelayStart", &Rules{}, client, delay, robots, false, true, true, nil},
		{"NoRobots", &Rules{Delay: time.Second}, client, delay, nil /*Robots*/, true, true, false, nil},
		{"NoDelayNoRobots", &Rules{}, client, nil /*Delay*/, nil /*Robots*/, false, false, false, nil},

		{
			"DoErr",
			&Rules{Fields: map[string]any{"doErr": testErr}},
			client,
			delay,
			robots,
			false,
			true,
			false,
			testErr,
		},
		{
			"RobotsErr",
			&Rules{Fields: map[string]any{"robotsErr": testErr}},
			client,
			nil, /*Delay*/
			robots,
			false,
			false,
			true,
			testErr,
		},

		{
			"Panic",
			&Rules{Fields: map[string]any{"doPanic": testErr}},
			client,
			nil, /*Delay*/
			nil, /*Robots*/
			false,
			false,
			false,
			testErr,
		},
	}

	for _, tt := range tests {
		c.Client = tt.Client
		c.Delay = tt.Delay
		c.RobotsTxt = tt.Robots

		t.Run(tt.Name, func(t *testing.T) {
			defer c.Clear()

			_, err := c.Do(tt.Rules)
			if (err != nil) && (tt.Err != nil) {
				if err.Error() != tt.Err.Error() {
					t.Fatal(err)
				}
				return

			} else if (err == nil) && (tt.Err == nil) {
				if delay.WaitUsed != tt.DelayWaitUsed {
					t.Fatal("Delay Wait")
				}

				if delay.DoneUsed != tt.DelayWaitUsed {
					t.Fatal("Delay Done")
				}

				if delay.StampUsed != tt.DelayStampUsed {
					t.Fatal("Delay Stamp")
				}

				if robots.IsAllowedUsed != tt.RobotsUsed {
					t.Fatal("RobotsTxt IsAllowed")
				}

				return
			}

			t.Fatal(err)
		})
	}

	// User-Agent
	t.Run("UserAgent", func(t *testing.T) {
		c := c
		c.Client = client

		tests := []struct {
			UserAgent, WantUserAgent string
		}{
			{"", DefaultUserAgent},
			{"    ", DefaultUserAgent},
			{"test/0.0.1", "test/0.0.1"},
		}

		rules, err := NewRules(map[string]any{"Header": nil})
		if err != nil {
			t.Fatal(err)
		}

		for _, tt := range tests {
			name := "(" + tt.UserAgent + " | " + tt.WantUserAgent + ")"
			rules.Header.Set("User-Agent", tt.UserAgent)

			t.Run(name, func(t *testing.T) {
				_, err := c.Do(rules)
				if err != nil {
					t.Fatal(err)
				}

				if rules.Header.Get("User-Agent") != tt.WantUserAgent {
					t.Fatal("not equal")
				}
			})
		}
	})
}

func TestColibriExtract(t *testing.T) {
	var (
		c      = New()
		client = &testClient{}
		parser = &testParser{}

		testErr = errors.New("Test Error")
	)
	c.RobotsTxt = &testRobots{}

	tests := []struct {
		Name   string
		Rules  *Rules
		Client HTTPClient
		Parser Parser
		Err    error
	}{
		{"OK", &Rules{}, client, parser, nil},

		{"ClientIsNil", &Rules{}, nil, parser, ErrClientIsNil},
		{"ParserIsNil", &Rules{}, client, nil, ErrParserIsNil},
		{"ParserIsNil2", &Rules{}, nil, nil, ErrParserIsNil},

		{
			"DoErr",
			&Rules{
				Fields: map[string]any{"doErr": testErr},
			},
			client,
			parser,
			testErr,
		},
		{
			"RobotsErr",
			&Rules{
				Fields: map[string]any{"robotsErr": testErr},
			},
			client,
			parser,
			testErr,
		},
		{
			"ParserErr",
			&Rules{
				Selectors: []*Selector{testSelector},
				Fields:    map[string]any{"parserErr": testErr},
			},
			client,
			parser,
			testErr,
		},

		{
			"Panic",
			&Rules{
				Selectors: []*Selector{testSelector},
				Fields:    map[string]any{"parserPanic": testErr},
			},
			client,
			parser,
			testErr,
		},
	}

	for _, tt := range tests {
		c.Client = tt.Client
		c.Parser = tt.Parser

		t.Run(tt.Name, func(t *testing.T) {
			_, _, err := c.Extract(tt.Rules)
			if (err != nil) && (tt.Err != nil) {
				if err.Error() != tt.Err.Error() {
					t.Fatal(err)
				}
				return

			} else if (err == nil) && (tt.Err == nil) {
				return
			}

			t.Fatal(err)
		})
	}
}

func TestNewRules(t *testing.T) {
	tests := []struct {
		Name      string
		RawRules  map[string]any
		WantRules *Rules
		ErrMap    map[string]any
	}{
		{"OK", testRawRules, testRules, nil},
		{"Nil", nil, &Rules{Fields: make(map[string]any)}, nil},
		{
			"NilSelectors",
			map[string]any{
				"URL":       "https://go.dev",
				"Selectors": nil,
			},
			&Rules{URL: mustNewURL("https://go.dev"), Fields: make(map[string]any)},
			nil,
		},

		{
			"InvalidSelectors",
			map[string]any{"Selectors": 21}, // ErrInvalidSelectors
			nil,
			map[string]any{
				"Selectors": ErrInvalidSelectors.Error(),
			},
		},
		{
			"InvalidSelector",
			map[string]any{
				"Selectors": map[string]any{
					"title": true, // ErrInvalidSelector
				},
			},
			nil,
			map[string]any{
				"Selectors": map[string]any{
					"title": ErrInvalidSelector.Error(),
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			rules, err := NewRules(tt.RawRules)
			defer ReleaseRules(rules)

			if (err != nil) && (tt.ErrMap != nil) {
				wantErr, _ := json.Marshal(tt.ErrMap)
				jsonErrs, _ := json.Marshal(err)

				if !reflect.DeepEqual(wantErr, jsonErrs) {
					t.Fatal(err)
				}
				return

			} else if (err == nil) && (tt.ErrMap == nil) {
				if !reflect.DeepEqual(rules, tt.WantRules) {
					t.Fatal("not equal")
				}
				return
			}

			t.Fatal(err)
		})
	}

	// without ConvFunc
	t.Run("WithoutConvFunc", func(t *testing.T) {
		_, err := NewRulesWithConvFunc(testRawRules, nil /*convFunc*/)
		if err == nil {
			t.Fatal("nil error")
		}
	})

	t.Run("ProcessRaw_NilFields", func(t *testing.T) {
		var (
			rawRules = map[string]any{"id": 21}

			rules = &Rules{}

			wantRules = &Rules{
				Fields: map[string]any{"id": 21},
			}
		)

		err := processRaw(rawRules, rules, DefaultConvFunc)
		if err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(rules, wantRules) {
			t.Fatal("not equal")
		}
	})

	t.Run("Clone", func(t *testing.T) {
		cRules := testRules.Clone()
		if !reflect.DeepEqual(cRules, testRules) {
			t.Fatal("not equal")
		}
	})
}

func TestRulesUnmarshalJSON(t *testing.T) {
	tests := []struct {
		Name      string
		RawRules  any
		WantRules *Rules
		AnErr     bool
	}{
		{"OK", testRawRules, testRules, false},

		{
			"Fail",
			map[string]any{
				"URL":             123,
				"IgnoreRobotsTxt": "error",
				"Selectors":       nil,
			},
			nil,
			true,
		},

		{
			"FailUnmarshal",
			"error",
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			data, err := json.Marshal(tt.RawRules)
			if err != nil {
				t.Fatal(err)
			}

			var newRules Rules
			err = json.Unmarshal(data, &newRules)
			if (err != nil && !tt.AnErr) || (err == nil && tt.AnErr) {
				t.Fatal(err)

			} else if (err == nil) && !tt.AnErr {
				if !reflect.DeepEqual(&newRules, tt.WantRules) {
					t.Fatalf("not equal")
				}
			}
		})
	}
}

func TestSelectorRules(t *testing.T) {
	t.Run("", func(t *testing.T) {
		selector := testSelector.Clone()
		selector.Fields = nil

		wantRules := testRules.Clone()
		wantRules.Method = ""
		wantRules.URL = nil
		wantRules.Fields = make(map[string]any)
		wantRules.Selectors = selector.Selectors

		rules := selector.Rules(testRules)
		if !reflect.DeepEqual(rules, wantRules) {
			t.Fatal("not equal")
		}
	})

	t.Run("", func(t *testing.T) {
		selector := testSelector.Clone()
		selector.Fields["Method"] = "POST"
		selector.Fields["Proxy"] = mustNewURL("")
		selector.Fields["Header"] = http.Header{"Accept": {"application/xml"}}
		selector.Fields["Timeout"] = 10 * time.Second
		selector.Fields["UseCookies"] = true
		selector.Fields["IgnoreRobotsTxt"] = false
		selector.Fields["Delay"] = 5 * time.Second

		wantRules := &Rules{
			Method:     "POST",
			Proxy:      mustNewURL(""),
			Header:     http.Header{"Accept": {"application/xml"}},
			Timeout:    10 * time.Second,
			UseCookies: true,
			Delay:      5 * time.Second,
			Selectors:  CloneSelectors(selector.Selectors),
			Fields:     make(map[string]any),
		}

		rules := selector.Rules(testRules)
		if !reflect.DeepEqual(rules, wantRules) {
			t.Fatal("not equal")
		}
	})

	t.Run("", func(t *testing.T) {
		selector := testSelector.Clone()
		selector.Fields["Method"] = "POST"
		selector.Fields["UseCookies"] = true
		selector.Fields["Delay"] = 5 * time.Second

		wantRules := &Rules{
			Method:          "POST",
			Proxy:           testRules.Proxy,
			Header:          testRules.Header,
			Timeout:         testRules.Timeout,
			UseCookies:      true,
			IgnoreRobotsTxt: testRules.IgnoreRobotsTxt,
			Delay:           5 * time.Second,
			Selectors:       CloneSelectors(selector.Selectors),
			Fields:          make(map[string]any),
		}

		rules := selector.Rules(testRules)
		if !reflect.DeepEqual(rules, wantRules) {
			t.Fatal("not equal")
		}
	})
}

func TestClear(t *testing.T) {
	t.Run("Colibri", func(t *testing.T) {
		var (
			c      = New()
			client = &testClient{}
			delay  = &testDelay{}
			robots = &testRobots{}
			parser = &testParser{}
		)

		c.Clear() // panic?

		c.Client = client
		c.Delay = delay
		c.RobotsTxt = robots
		c.Parser = parser

		if (c.Client == nil) || (c.Delay == nil) ||
			(c.RobotsTxt == nil) || (c.Parser == nil) {
			t.Fatal("nil field")
		}

		c.Clear()

		if !(client.ClearUsed && delay.ClearUsed &&
			robots.ClearUsed && parser.ClearUsed) {
			t.Fatal("field not clear")
		}
	})

	t.Run("Rules", func(t *testing.T) {
		rules, err := NewRules(testRawRules)
		if err != nil {
			t.Fatal(err)
		}

		rules.Clear()

		want := &Rules{Fields: map[string]any{}}
		if !reflect.DeepEqual(rules, want) {
			t.Fatal("Uncleaned")
		}

		// panic?
		rules2 := &Rules{}
		ReleaseRules(rules2)
	})

	t.Run("Selector", func(t *testing.T) {
		selector, err := newSelector("head", testRawSelector, DefaultConvFunc)
		if err != nil {
			t.Fatal(err)
		}

		selector.Clear()

		want := &Selector{Fields: map[string]any{}}
		if !reflect.DeepEqual(selector, want) {
			t.Fatal("Uncleaned")
		}

		// panic?
		selector2 := &Selector{}
		ReleaseSelector(selector2)
	})
}

func TestErrs(t *testing.T) {
	var (
		err1 = errors.New("err 1")
		err2 = errors.New("err 2")
		err3 = errors.New("err 3")
	)

	t.Run("AddError", func(t *testing.T) {
		errs := &Errs{}
		AddError(errs, "err1", err1)

		err, _ := errs.Get("err1")
		if !errors.Is(err, err1) {
			t.Fatal(err)
		}
	})

	t.Run("AddError_Nil", func(t *testing.T) {
		errs := AddError(nil, "err2", err2)

		err, _ := errs.(*Errs).Get("err2")
		if !errors.Is(err, err2) {
			t.Fatal(err)
		}
	})

	t.Run("AddError_Error", func(t *testing.T) {
		errs := AddError(err2, "err3", err3)

		err, _ := errs.(*Errs).Get("#")
		if !errors.Is(err, err2) {
			t.Fatal(err)
		}

		err, _ = errs.(*Errs).Get("err3")
		if !errors.Is(err, err3) {
			t.Fatal(err)
		}
	})

	t.Run("AddError_IgnoreErr", func(t *testing.T) {
		var errs error
		errs = AddError(errs, "err1", nil) // ignore
		errs = AddError(errs, "", err1)    // ignore

		if errs != nil {
			t.Fatal(errs)
		}
	})

	subErr := AddError(nil, "err3", err3)

	errs := &Errs{}
	errs.Add("err1", err1).
		Add("err2", err2).
		Add("sub", subErr).
		Add("err1", err1)

	errs.Add("err4", nil).Add("", err3) // ignore

	want := map[string]any{
		"err1":   "err 1",
		"err1#1": "err 1",
		"sub": map[string]any{
			"err3": "err 3",
		},
		"err2": "err 2",
	}

	var result map[string]any

	if err := json.Unmarshal([]byte(errs.Error()), &result); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result, want) {
		t.Fatal("not equal", result)
	}
}

func TestDefaultConvFunc(t *testing.T) {
	var emptySelectorSlice []*Selector

	tests := []struct {
		Key               string
		Input, WantOutput any
		AnErr             bool
	}{
		// URL
		{KeyURL, "/", mustNewURL("/"), false},
		{KeyProxy, "", mustNewURL(""), false},

		{KeyURL, nil, nil, true},
		{KeyProxy, true, nil, true},

		// Bool
		{KeyIgnoreRobotsTxt, "true", true, false /*AnErr*/},
		{KeyFollow, 0, false, false /*AnErr*/},
		{KeyUseCookies, uint(1), true, false /*AnErr*/},
		{KeyAll, 1.5, true, false /*AnErr*/},
		{KeyIgnoreRobotsTxt, "f", false, false /*AnErr*/},
		{KeyFollow, nil, false, false /*AnErr*/},

		{KeyUseCookies, []byte{}, false, true /*AnErr*/},
		{KeyAll, "error", false, true /*AnErr*/},

		// Duration
		{KeyTimeout, nil, time.Duration(0), false},
		{KeyDelay, "3m", 3 * time.Minute, false},
		{KeyTimeout, 2, 2 * time.Millisecond, false},
		{KeyDelay, uint(1), 1 * time.Millisecond, false},
		{KeyTimeout, 1.5, 1500000 * time.Nanosecond, false},

		{KeyDelay, "error", time.Duration(0), true},
		{KeyTimeout, []byte{}, time.Duration(0), true},

		// Header
		{KeyHeader, nil, http.Header{}, false},
		{
			KeyHeader,
			map[string]any{
				"User-Agent": "test/0.1",
				"Accept":     []string{"application/json", "application/xml"},
			},
			http.Header{
				"User-Agent": {"test/0.1"},
				"Accept":     {"application/json", "application/xml"},
			},
			false,
		},

		{KeyHeader, 123, http.Header{}, true},
		{
			KeyHeader,
			map[any]any{123: "test/0.1"},
			nil,
			true,
		},
		{
			KeyHeader,
			map[string]any{"User-Agent": 123},
			nil,
			true,
		},

		// Selectors
		{
			KeySelectors,
			map[string]any{"head": testRawSelector},
			[]*Selector{testSelector},
			false,
		},
		{KeySelectors, nil, emptySelectorSlice, false},

		{KeySelectors, []byte{}, emptySelectorSlice, true /*ErrInvalidSelectors*/},
		{
			KeySelectors,
			map[string]any{"title": []byte{}},
			emptySelectorSlice,
			true, /*ErrInvalidSelector*/
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.Key, func(t *testing.T) {
			t.Parallel()

			output, err := DefaultConvFunc(tt.Key, tt.Input)
			if (err != nil && !tt.AnErr) || (err == nil && tt.AnErr) {
				t.Fatal(err)

			} else if (err == nil) && !tt.AnErr {
				if !reflect.DeepEqual(output, tt.WantOutput) {
					t.Fatalf("not equal")
				}
			}
		})
	}
}

func BenchmarkNewRules(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		rules, err := NewRules(testRawRules)
		if err != nil {
			b.Fatal(err)
		}
		ReleaseRules(rules)
	}
}

var (
	testRawSelector = map[string]any{
		"Expr":   "//head",
		"Type":   "xpath",
		"All":    "false",
		"Follow": 0,

		"Selectors": map[string]any{
			"title": "//title",
			"":      "//script/@src", // ignore
			"media": "",              // ignore
		},

		"required": true,
	}

	testRawRules = map[string]any{
		"Method": "GET",
		"URL":    "https://pkg.go.dev",
		"Proxy":  "http://proxy-url.com:8080",
		"Header": map[string]string{
			"User-Agent": "test/0.0.1",
		},
		"Timeout": "2s",

		"UseCookies":      "true",
		"IgnoreRobotsTxt": true,
		"Delay":           1,

		"Selectors": map[string]any{
			"head": testRawSelector,

			"body": nil, // ignore
		},

		"id":    float64(123), // UnmarshalJSON
		"token": "T456",
	}

	testSelector = &Selector{
		Name:   "head",
		Expr:   "//head",
		Type:   "xpath",
		All:    false,
		Follow: false,
		Selectors: []*Selector{
			{Name: "title", Expr: "//title", Fields: make(map[string]any)},
		},
		Fields: map[string]any{
			"required": true,
		},
	}

	testRules = &Rules{
		Method:          "GET",
		URL:             mustNewURL("https://pkg.go.dev"),
		Proxy:           mustNewURL("http://proxy-url.com:8080"),
		Header:          http.Header{"User-Agent": {"test/0.0.1"}},
		Timeout:         2 * time.Second,
		UseCookies:      true,
		IgnoreRobotsTxt: true,
		Delay:           1 * time.Millisecond,

		Selectors: []*Selector{testSelector},

		Fields: map[string]any{
			"id":    float64(123), // UnmarshalJSON
			"token": "T456",
		},
	}
)

type testResp struct{}

func (resp *testResp) URL() *url.URL       { return nil }
func (resp *testResp) StatusCode() int     { return 500 }
func (resp *testResp) Header() http.Header { return nil }
func (resp *testResp) Body() io.ReadCloser { return nil }
func (resp *testResp) Do(_ *Rules) (Response, error) {
	return &testResp{}, nil
}
func (resp *testResp) Extract(_ *Rules) (Response, map[string]any, error) {
	return resp, make(map[string]any), nil
}

type testClient struct {
	ClearUsed bool
}

func (c *testClient) Do(_ *Colibri, rules *Rules) (Response, error) {
	if err := rules.Fields["doErr"]; err != nil {
		return nil, err.(error)
	} else if v := rules.Fields["doPanic"]; v != nil {
		panic(v)
	}
	return &testResp{}, nil
}
func (c *testClient) Clear() { c.ClearUsed = true }

type testDelay struct {
	WaitUsed, DoneUsed, StampUsed, ClearUsed bool
}

func (d *testDelay) Wait(_ *url.URL, _ time.Duration) { d.WaitUsed = true }
func (d *testDelay) Done(_ *url.URL)                  { d.DoneUsed = true }
func (d *testDelay) Stamp(_ *url.URL)                 { d.StampUsed = true }
func (d *testDelay) Clear() {
	d.ClearUsed = true
	d.WaitUsed = false
	d.DoneUsed = false
	d.StampUsed = false
}

type testRobots struct {
	IsAllowedUsed, ClearUsed bool
}

func (r *testRobots) IsAllowed(_ *Colibri, rules *Rules) error {
	r.IsAllowedUsed = true
	err := rules.Fields["robotsErr"]
	if err != nil {
		return err.(error)
	}
	return nil
}
func (r *testRobots) Clear() {
	r.ClearUsed = true
	r.IsAllowedUsed = false
}

type testParser struct {
	ParseUsed, ClearUsed bool
}

func (p *testParser) Match(_ string) bool { return false }
func (p *testParser) Parse(rules *Rules, _ Response) (map[string]any, error) {
	p.ParseUsed = true

	if err := rules.Fields["parserErr"]; err != nil {
		return nil, err.(error)
	} else if v := rules.Fields["parserPanic"]; v != nil {
		panic(v)
	}
	return make(map[string]any), nil
}
func (p *testParser) Clear() {
	p.ClearUsed = true
}

func mustNewURL(rawURL string) *url.URL {
	u, _ := url.Parse(rawURL)
	return u
}
