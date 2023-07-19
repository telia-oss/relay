package relay

import (
	"sync"
	"time"
)

type RequestInfo struct {
	Timestamp time.Time
	Success   bool
}

type Counters struct {
	mutex            sync.RWMutex
	Window           []RequestInfo
	WindowWidth      time.Duration
	HalfOpenRequests uint32
}

func (c *Counters) Add(success bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	minTime := now.Add(-c.WindowWidth)
	for len(c.Window) > 0 && c.Window[0].Timestamp.Before(minTime) {
		c.Window = c.Window[1:]
	}

	c.Window = append(c.Window, RequestInfo{Timestamp: now, Success: success})
}

func (c *Counters) FailuresAndSuccessesCount(r *Relay) (int, int) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	failures := 0
	successes := 0
	for _, req := range r.counters.Window {
		if !req.Success {
			failures++
		} else {
			successes++
		}
	}

	return failures, successes

}

func (c *Counters) clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.Window = []RequestInfo{}
	c.HalfOpenRequests = 0
}
