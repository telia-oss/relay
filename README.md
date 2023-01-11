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
        *relay.NewConfig().WithCoolDown(20).WithFailuresThreshold(20).WithHalfOpenRequestsQuota(30).WithSuccessesThreshold(10)))

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
| GetRelay      | Returns circute breaker by name |