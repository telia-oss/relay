package relay

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	// Assert transition from Closed to Open state
	for i := 0; i < int(*relay.config.FailuresThreshold); i++ {
		//  default FailuresThreshold = 10
		fail(relay) // fail 10 times in Closed state.
	}
	assert.Equal(t, Open, relay.State())

	// Assert transition from Open to Half-Open state.
	fail(relay)
	time.Sleep(time.Duration(*relay.config.CoolDown) * time.Second)
	fail(relay)
	assert.Equal(t, HalfOpen, relay.State())

	// Assert transition from Half-Open to Closed state
	for i := 0; i < int(*relay.config.SuccessesThreshold); i++ {
		success(relay) // succeed 3 times
	}
	assert.Equal(t, Closed, relay.State())

	// Assert transition from Half-Open to Open state
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

func TestCustomErrorCodes(t *testing.T) {
	relay := Must(New("customCodes",
		*NewConfig().WithGrpcCodes([]codes.Code{codes.Internal})))

	assert.Equal(t, []codes.Code{codes.Internal}, *relay.config.GrpcCodes)

	// Assert no transition from Closed to Open state when returning errors are not gRPC errors
	for i := 0; i < int(*relay.config.FailuresThreshold); i++ {
		//  default FailuresThreshold = 10
		fail(relay) // fail 10 times in Closed state.
	}
	assert.Equal(t, Closed, relay.State())

	// Assert transition from Closed to Open state when the right gRPC error occur
	for i := 0; i < int(*relay.config.FailuresThreshold); i++ {
		//  default FailuresThreshold = 10
		grpcFail(relay, codes.Internal)
	}
	assert.Equal(t, Open, relay.State())

	// Assert no transition from Closed to Open state when the wrong gRPC error occur
	relay.setState(Closed)
	for i := 0; i < int(*relay.config.FailuresThreshold); i++ {
		//  default FailuresThreshold = 10
		grpcFail(relay, codes.InvalidArgument)
	}
	assert.Equal(t, Closed, relay.State())
}

type stateMock struct {
	mock.Mock
}

func (s *stateMock) OnStateChange(name string, from State, to State) {

	s.Called(name, from, to)

	fmt.Printf("Transiting from state: %v to state: %v", from, to)
}

func TestOnStateChange(t *testing.T) {

	mockedState := stateMock{}
	mockedState.On("OnStateChange", "custom", Closed, Open)

	relay := Must(New("custom",
		*NewConfig().WithOnStateChange(mockedState.OnStateChange)))
	for i := 0; i < int(*relay.config.FailuresThreshold); i++ {
		fail(relay)
	}

	mockedState.AssertCalled(t, "OnStateChange", "custom", Closed, Open)

}

func fail(relay *Relay) error {
	relay.Relay(func() (interface{}, error) { return nil, errors.New("Test error") })
	return nil
}

func grpcFail(relay *Relay, code codes.Code) error {
	err := status.Errorf(code, "Internale server error")
	relay.Relay(func() (interface{}, error) { return nil, err })
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
