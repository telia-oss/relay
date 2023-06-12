package relay

import (
	"errors"
	"sync"
	"time"

	"google.golang.org/grpc/status"
)

type Relay struct {
	config   Config
	state    State
	mutex    sync.RWMutex
	expiry   time.Time
	counters Counters
}

var relays = make(map[string]*Relay)

func New(name string, confs ...Config) (*Relay, error) {
	if name == "" {
		return nil, errors.New("relay name should be set")
	}

	relay := new(Relay)

	if len(confs) == 0 {
		relay.setState(Closed)

		relay.config = Config{
			Name: &name,
		}
		relay.config.WithCoolDown(10)
		relay.config.WithSuccessesThreshold(3)
		relay.config.WithFailuresThreshold(10)
		relay.config.WithHalfOpenRequestsQuota(10)

		relay.counters = Counters{
			Window:      []RequestInfo{},
			WindowWidth: time.Duration(*relay.config.CoolDown) * time.Second,
		}
	}

	for _, conf := range confs {
		relay.config = conf

		relay.setState(Closed)

		relay.config.Name = &name

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

		relay.counters = Counters{
			Window:      []RequestInfo{},
			WindowWidth: time.Duration(*relay.config.CoolDown) * time.Second,
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

func Relays() map[string]*Relay {
	return relays
}

var _ *Relay = (*Relay)(nil)

func (r *Relay) Get(name string) (*Relay, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	relay, exists := relays[name]
	if !exists {
		return nil, errors.New("relay not found")
	}

	return relay, nil
}

func (r *Relay) State() State {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.state
}

func (r *Relay) Config() Config {
	return r.config
}

// Request wrapper function
func (r *Relay) Relay(req func() (interface{}, error)) (interface{}, error) {
	// failures, successes := r.counters.FailuresAndSuccessesCount(r)
	switch r.State() {
	case Open:
		now := time.Now()
		/*
		* if the circute breaker state expired
		* excute the function OnStateChange if any fucntion is attached
		* set state to half open
		 */
		if r.expiry.Before(now) {
			if r.config.OnStateChange != nil {
				r.config.OnStateChange(*r.config.Name, r.state, HalfOpen)
			}
			r.setState(HalfOpen)
		}
		// if the circute breaker not expired return circute breaker open error
		return nil, errors.New("this service circute is open")
	case HalfOpen:
		if r.counters.HalfOpenRequests > *r.config.HalfOpenRequestsQuota {
			return nil, errors.New("half open request quota execceded")
		}
		r.counters.HalfOpenRequests++

		result, err := req()

		if err != nil {
			result, err = r.examineError(err, func() (interface{}, error) {
				if r.config.OnStateChange != nil {
					r.config.OnStateChange(*r.config.Name, r.state, Open)
				}
				r.setState(Open)
				// reset the request countres
				r.counters.clear()
				return nil, err
			})
			return result, err
		}
		r.counters.Add(true)
		_, successes := r.counters.FailuresAndSuccessesCount(r)
		if uint32(successes) >= *r.config.SuccessesThreshold {
			if r.config.OnStateChange != nil {
				r.config.OnStateChange(*r.config.Name, r.state, Closed)
			}
			r.setState(Closed)
		}
		return result, err
	}
	result, err := req()
	if err != nil {
		result, err = r.examineError(err, func() (interface{}, error) {
			r.counters.Add(false)
			failures, _ := r.counters.FailuresAndSuccessesCount(r)
			if uint32(failures) >= *r.config.FailuresThreshold {
				if r.config.OnStateChange != nil {
					r.config.OnStateChange(*r.config.Name, r.state, Open)
				}
				r.setState(Open)
				r.expiry = time.Now().Add(time.Duration(*r.config.CoolDown) * time.Second)
			}
			return nil, err
		})
		return result, err
	}
	// reset counters
	r.counters.clear()

	return result, err
}

func (r *Relay) setState(state State) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.state = state
}

func add(r *Relay) {
	relays[*r.config.Name] = r
}

func (r *Relay) examineError(err error, callback func() (interface{}, error)) (interface{}, error) {
	if r.config.GrpcCodes == nil || len(*r.config.GrpcCodes) == 0 {
		return callback()
	}
	for _, errorCode := range *r.config.GrpcCodes {
		if grpcError, ok := status.FromError(err); ok {
			if grpcError.Code() == errorCode {
				return callback()
			}
		}
	}
	return nil, err
}
