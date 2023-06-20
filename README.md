Relay is a gRPC circuit breaker and Protoc plugin designed in Go. It provides fault-tolerance by monitoring gRPC calls for failures and preventing further system calls once a failure threshold is reached.

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
| CoolDown              | Duration, in seconds, before transitioning from the OPEN state to the HALF-OPEN state. Additionally, it determines the sliding window size for request counting. When a request exceeds this 'CoolDown' duration, it's removed from the sliding window and won't influence future state transition decisions made by the Relay |
| SuccessesThreshold    | Number of successful requsts to transit  from HALF-OPEN  to Closed state |
| FailuresThreshold     | Number of failed requsts to transit from CLOSED to OPEN state |
| HalfOpenRequestsQuota | Number of requsts allowed during HALF-OPEN  state |
| GrpcCodes             | gRPC error codes, if no code is passed Relay will count every error |
| OnStateChange         | A function execute when transiting from a state to another state |

## gRPC Service Code Generation with protoc-gen-relay

When working with numerous services in your gRPC server, you might find it challenging to manually create Relay configurations to register Relay circuit breakers for each service. The `protoc-gen-relay` is a `protoc` plugin can automate this task for you.

This plugin helps you to generate the necessary Go code to register Relay circuit breakers for each of your services, saving you from the repetitive task and ensuring that all the services adhere to the same configuration pattern.

Detailed instructions for installing, building, and using the `protoc-gen-relay` plugin can be found in the plugin's [README file](https://github.com/telia-oss/relay/tree/main/pkg/protoc-gen-relay/README.md).

## Contributing

We appreciate and welcome all contributions to the Relay project, no matter how big or small. Whether you've found a typo that needs correcting or you're considering a major addition to the codebase, your input is incredibly valuable to us.

To contribute, feel free to open an issue or submit a pull request on GitHub. Your involvement helps us build a robust and dependable tool for everyone.

### Testing

Our commitment to robustness and scalability is strongly backed by our rigorous approach to testing. We believe that a well-tested codebase is fundamental to building reliable backend connectors. We work hard to maintain a stable and comprehensive test suite that validates the functionality of our code.

When contributing, please ensure to write tests for your code and confirm that your changes do not disrupt the existing tests. This helps maintain the quality and integrity of our project. Thank you for helping us in this mission.

### Commit message

To keep neat commit history please follow the commit messages structure:

- Feat
- Fix
- Chore
- Doc

Thank you!