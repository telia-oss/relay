package relay

import (
	"sync"
	"time"
)

type Relay struct {
	config   Config
	state    State
	mutex    sync.RWMutex
	expiry   time.Time
	counters Counters
}

type Counters struct {
	Successes uint32
	Failures  uint32
	Requests  uint32
}
