package relay

import (
	"errors"
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

var relays []*Relay

type Counters struct {
	Successes uint32
	Failures  uint32
	Requests  uint32
}

func New(name string, confs ...Config) (*Relay, error) {
	if name == "" {
		return nil, errors.New("Relay name should be set")
	}

	relay := new(Relay)

	if len(confs) == 0 {
		relay.config = Config{
			Name: &name,
		}
		relay.setState(Closed)
		relay.config.WithCoolDown(10)
		relay.config.WithSuccessesThreshold(3)
		relay.config.WithFailuresThreshold(10)
		relay.config.WithHalfOpenRequestsQuota(10)
	}

	for _, conf := range confs {
		relay.config = conf
		relay.config.Name = &name
		relay.setState(Closed)
		if relay.config.CoolDown == nil {
			relay.config.WithCoolDown(5)
		}
		if relay.config.CoolDown == nil {
			relay.config.WithCoolDown(5)
		}
		if relay.config.SuccessesThreshold == nil {
			relay.config.WithSuccessesThreshold(3)
		}
		if relay.config.FailuresThreshold == nil {
			relay.config.WithFailuresThreshold(10)
		}
		if relay.config.HalfOpenRequestsQuota == nil {
			relay.config.WithHalfOpenRequestsQuota(10)
		}

	}
	add(relay)
	return relay, nil
}

func Must(relay *Relay, err error) *Relay {
	if err != nil {
		panic(err)
	}

	return relay
}

func Relays() []*Relay {
	return relays
}

func GetRelay(name string) *Relay {
	for _, relay := range relays {
		if *relay.Config().Name == name {
			return relay
		}
	}
	return nil
}

func (r *Relay) State() State {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.state
}

func (r *Relay) Config() Config {
	return r.config
}

func add(r *Relay) {
	relays = append(relays, r)
}

var _ *Relay = (*Relay)(nil)

// Request wrapper function
func (r *Relay) Relay(req func() (interface{}, error)) (interface{}, error) {
	switch r.State() {
	case Open:
		now := time.Now()
		if r.expiry.Before(now) {
			r.setState(HalfOpen)
		}
		return nil, errors.New("this service circute is open")
	case HalfOpen:
		if r.counters.Requests > *r.config.HalfOpenRequestsQuota {
			return nil, errors.New("half open request quota execceded")
		}
		r.counters.Requests++

		result, err := req()

		if err != nil {
			r.setState(Open)
			// reset the request countres
			r.counters.clear()
			return nil, err
		}
		r.counters.Successes++
		if r.counters.Successes >= *r.config.SuccessesThreshold {
			r.setState(Closed)
		}
		return result, err
	}
	result, err := req()
	if err != nil {
		r.counters.Failures++
		if r.counters.Failures >= *r.config.FailuresThreshold {
			r.setState(Open)
			r.expiry = time.Now().Add(time.Duration(*r.config.CoolDown) * time.Second)
		}
		return nil, err
	}
	// reset Failures counter
	r.counters.clear()

	return result, err
}

func (c *Counters) clear() {
	c.Successes = 0
	c.Failures = 0
	c.Requests = 0
}

func (r *Relay) setState(state State) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	r.state = state
}
