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
    circuteBreaker := relay.Must(relay.New("custom", 
        *relay.NewConfig().WithGrpcCodes([]codes.Code{codes.Internal}).WithOnStateChange(func(name string, from relay.State, to relay.State) {
            fmt.Printf("State of cb %s changed from %v to %v \n", name, from, to)
        })))


    circuteBreaker.Relay(func() (interface{}, error) { return nil, errors.New("Test error") })
}
```

### Create Circute breaker with custom gRPC errors

When defining `WithGrpcCodes` relay will ignore all erorrs which doesn't belong to the error slice passed to `WithGrpcCodes`
```
import ( 
    "github.com/telia-oss/relay"
    "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
    circuteBreaker := relay.Must(relay.New("customCodes",
	    *relay.NewConfig().WithGrpcCodes([]codes.Code{codes.Internal})))

    // Internal error will be counted

    circuteBreaker.Relay(func() (interface{}, error) { return nil, status.Errorf(codes.Internal, "Internale server error") })

    // InvalidArgument error will be ignored
    circuteBreaker.Relay(func() (interface{}, error) { return nil, status.Errorf(codes.InvalidArgument, "Invalid argument") })
}
```

## Methods

| Name          | Description   |
| ------------- | ------------- | 
| Relay         | Execute circute breaker on function / handler |
| Relays        | Returns all registered circute breakers |
| Get           | Returns circute breaker by name |
| State         | Retruns a circute breaker state|
| Config        | Returns a circute breaker config

## Configurations

| Name          | Description   |
| ------------- | ------------- | 
| Name                  | Name of the circute  breaker|
| CoolDown              | Time in seconds to transit from OPEN to HALF-OPEN state |
| SuccessesThreshold    | Number of successful requsts to transit  from HALF-OPEN  to Closed state |
| FailuresThreshold     | Number of failed requsts to transit from CLOSED to OPEN state |
| HalfOpenRequestsQuota | Number of requsts allowed during HALF-OPEN  state |
| GrpcCodes             | gRPC error codes, if no code is passed Relay will count every error |
| OnStateChange         | A function execute when transiting from a state to another state |