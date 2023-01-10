gRPC Circuit breaker implementation in Go.

Note: this package is under development and has not yet been tested in a production environment.

## Usage

### Create Circute breaker with default configuration
```
import "github.com/telia-oss/relay"

func main() {
    circuteBreaker := relay.Must(relay.New("default"))

    circuteBreaker.Relay(func() (interface{}, error) { return nil, errors.New("Test error") })
}
```

### Create Circute breaker with custom configuration
```
import "github.com/telia-oss/relay"

func main() {
    circuteBreaker := relay.Must(relay.New("custom", *relay.NewConfig().WithCoolDown(20).WithFailuresThreshold(20).WithHalfOpenRequestsQuota(30).WithSuccessesThreshold(10)))

    circuteBreaker.Relay(func() (interface{}, error) { return nil, errors.New("Test error") })
}
```

## Methods

| Name          | 
| ------------- | 
| Relay         |
| Relays        | 
| GetRelay      | 