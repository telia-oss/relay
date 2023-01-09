package relay

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefualtConfig(t *testing.T) {
	relay := Must(New("default"))

	assert.Equal(t, "default", *relay.config.Name)
	assert.Equal(t, uint32(10), *relay.config.CoolDown)
	assert.Equal(t, uint32(3), *relay.config.SuccessesThreshold)
	assert.Equal(t, uint32(10), *relay.config.FailuresThreshold)
	assert.Equal(t, uint32(10), *relay.config.HalfOpenRequestsQuota)
	assert.Nil(t, relay.config.OnStateChange)
}

func TestCustomConfig(t *testing.T) {
	relay := Must(New("custom",
		*NewConfig().WithCoolDown(20).WithSuccessesThreshold(5).WithFailuresThreshold(20).WithHalfOpenRequestsQuota(20)))

	assert.Equal(t, "custom", *relay.config.Name)
	assert.Equal(t, uint32(20), *relay.config.CoolDown)
	assert.Equal(t, uint32(5), *relay.config.SuccessesThreshold)
	assert.Equal(t, uint32(20), *relay.config.FailuresThreshold)
	assert.Equal(t, uint32(20), *relay.config.HalfOpenRequestsQuota)
	assert.Nil(t, relay.config.OnStateChange)
}

func TestStateTransitionWithDefaultConfig(t *testing.T) {
	relay := Must(New("default"))

	// Assert transation from Closed to Open state
	for i := 0; i < int(*relay.config.FailuresThreshold); i++ {
		//  default FailuresThreshold = 10
		fail(relay) // fail 10 times in Closed state.
	}
	assert.Equal(t, Open, relay.State())

	// Assert transation from Open to Half-Open state.
	fail(relay)
	time.Sleep(time.Duration(*relay.config.CoolDown) * time.Second)
	fail(relay)
	assert.Equal(t, HalfOpen, relay.State())

	// Assert transation from Half-Open to Closed state
	for i := 0; i < int(*relay.config.SuccessesThreshold); i++ {
		success(relay) // succeed 3 times
	}
	assert.Equal(t, Closed, relay.State())

	// Assert transation from Half-Open to Open state
	relay.setState(HalfOpen)
	fail(relay)
	assert.Equal(t, Open, relay.State())

}

func TestHalfOpenRequestsQuota(t *testing.T) {
	relay := Must(New("default"))
	relay.setState(HalfOpen)

	ch := make(chan error)

	go successAndCollectRejections(relay, ch)

	numberOfRejectRequests := 0
	for err := range ch {
		if err.Error() == "half open request quota execceded" {
			numberOfRejectRequests++
		}
	}

	// ~ execute 20 request while Relay in half-open state
	// Result might vary depending on the timing of the concurent requests calling Relay
	// However ~10 requests should get (half open request quota execceded)
	assert.Greater(t, numberOfRejectRequests, 0)
	assert.Equal(t, Closed, relay.State())
}

func fail(relay *Relay) error {
	relay.Relay(func() (interface{}, error) { return nil, errors.New("Test error") })
	return nil
}

func success(relay *Relay) error {
	relay.Relay(func() (interface{}, error) { return nil, nil })
	return nil
}

func successAndCollectRejections(relay *Relay, ch chan error) {
	var wg sync.WaitGroup
	for i := 0; i <= 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := relay.Relay(func() (interface{}, error) {
				time.Sleep(50000)
				return nil, nil
			})
			if err != nil {
				ch <- err
			}
		}()

	}
	wg.Wait()
	close(ch)
}
