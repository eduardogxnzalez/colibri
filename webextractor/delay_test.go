package webextractor

import (
	"net/url"
	"testing"
	"time"
)

func TestReqDelay(t *testing.T) {
	tests := []struct {
		URL           *url.URL
		DelayDuration time.Duration
	}{
		{mustNewURL("https://pkg.go.dev"), 50 * time.Millisecond},
		{mustNewURL("https://pkg.go.dev"), 0},
		{mustNewURL("https://pkg.go.dev"), 25 * time.Millisecond},

		{mustNewURL("https://go.dev"), 25 * time.Millisecond},
	}

	delay := NewReqDelay()
	for _, tt := range tests {
		var (
			tt    = tt
			name  = "Host(" + tt.URL.Host + ")"
			first = !delay.visit(tt.URL)
		)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			start := time.Now()

			delay.Wait(tt.URL, tt.DelayDuration)
			time.Sleep(3 * time.Millisecond)
			delay.Done(tt.URL)
			delay.Stamp(tt.URL)

			end := time.Since(start)

			if !first && (end < tt.DelayDuration) {
				t.Fatal("Delay is not expected")
			}
		})
	}
}

func TestReqClear(t *testing.T) {
	var (
		delay    = NewReqDelay()
		u        = mustNewURL("https://pkg.go.dev")
		duration = 1 * time.Millisecond
	)

	if delay.visit(u) {
		t.Fatal("URL visited")
	}

	delay.Wait(u, duration)
	delay.Done(u)
	delay.Stamp(u)

	if !delay.visit(u) {
		t.Fatal("URL not visit")
	}

	delay.Clear()

	if delay.visit(u) {
		t.Fatal("Uncleaned")
	}
}
