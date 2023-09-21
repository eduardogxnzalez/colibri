package webextractor

import (
	"net/url"
	"sync"
	"time"
)

// ReqDelay manages the delay between each HTTP request.
// See the colibri.Delay interface.
type ReqDelay struct {
	rw        sync.RWMutex
	timestamp map[string]int64
	done      map[string]chan struct{}
}

// NewReqDelay returns a new ReqDelay structure.
func NewReqDelay() *ReqDelay {
	return &ReqDelay{
		timestamp: make(map[string]int64),
		done:      make(map[string]chan struct{}),
	}
}

func (rd *ReqDelay) Wait(u *url.URL, duration time.Duration) {
	rd.rw.RLock()
	ch, ok := rd.done[u.Host]
	rd.rw.RUnlock()

	if ok {
		<-ch

	} else {
		rd.rw.Lock()
		rd.done[u.Host] = make(chan struct{}, 1)
		rd.rw.Unlock()
	}

	rd.rw.RLock()
	timestamp, ok := rd.timestamp[u.Host]
	rd.rw.RUnlock()

	if ok {
		diff := duration.Milliseconds() - (time.Now().UnixMilli() - timestamp)
		if diff > 0 {
			time.Sleep(time.Duration(diff) * time.Millisecond)
		}
	}
}

func (rd *ReqDelay) Done(u *url.URL) {
	rd.rw.Lock()
	select {
	case rd.done[u.Host] <- struct{}{}:
	default:
	}
	rd.rw.Unlock()
}

func (rd *ReqDelay) Stamp(u *url.URL) {
	rd.rw.Lock()
	rd.timestamp[u.Host] = time.Now().UnixMilli()
	rd.rw.Unlock()
}

func (rd *ReqDelay) Clear() {
	rd.rw.Lock()
	clear(rd.timestamp)

	for host := range rd.done {
		close(rd.done[host])
		delete(rd.done, host)
	}
	rd.rw.Unlock()
}

func (rd *ReqDelay) visit(u *url.URL) bool {
	rd.rw.RLock()
	_, ok := rd.timestamp[u.Host]
	rd.rw.RUnlock()
	return ok
}
