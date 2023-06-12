package relay

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {

	counters := &Counters{
		Window:      make([]RequestInfo, 0),
		WindowWidth: 1 * time.Second,
	}

	assert.Equal(t, 0, len(counters.Window))
	counters.Add(true)
	assert.Equal(t, 1, len(counters.Window))
}

func TestFailuresAndSuccessesCount(t *testing.T) {
	counters := &Counters{
		Window:      make([]RequestInfo, 0),
		WindowWidth: 5 * time.Second,
	}
	relay := &Relay{
		counters: *counters,
	}

	// Add 2 successes and 1 failure.
	relay.counters.Add(true)
	relay.counters.Add(true)
	relay.counters.Add(false)

	failures, successes := counters.FailuresAndSuccessesCount(relay)

	assert.Equal(t, 1, failures)
	assert.Equal(t, 2, successes)
}

func TestClear(t *testing.T) {
	counters := &Counters{
		Window:           make([]RequestInfo, 0),
		WindowWidth:      5 * time.Second,
		HalfOpenRequests: 5,
	}

	counters.Add(true)
	counters.clear()

	assert.Equal(t, 0, len(counters.Window))
	assert.Equal(t, 0, int(counters.HalfOpenRequests))
}
