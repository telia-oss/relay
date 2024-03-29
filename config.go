package relay

import "google.golang.org/grpc/codes"

type State int

const (
	Closed State = iota + 1
	HalfOpen
	Open
)

type Config struct {
	Name *string
	// time in seconds to move from open to half-open
	CoolDown *uint32
	// number of successful requests to move  from half-open to closed
	SuccessesThreshold *uint32
	// number of failed requests to move from closed to open
	FailuresThreshold *uint32
	// number of requests allowed in half-open state
	HalfOpenRequestsQuota *uint32
	OnStateChange         func(name string, from State, to State)
	// gRPC error codes, if no code is passed Relay will count every error.
	GrpcCodes *[]codes.Code
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) WithCoolDown(period uint32) *Config {
	c.CoolDown = &period
	return c
}

func (c *Config) WithSuccessesThreshold(count uint32) *Config {
	c.SuccessesThreshold = &count
	return c
}

func (c *Config) WithFailuresThreshold(count uint32) *Config {
	c.FailuresThreshold = &count
	return c
}

func (c *Config) WithHalfOpenRequestsQuota(count uint32) *Config {
	c.HalfOpenRequestsQuota = &count
	return c
}

func (c *Config) WithOnStateChange(onStateChange func(name string, from State, to State)) *Config {
	c.OnStateChange = onStateChange
	return c
}

func (c *Config) WithGrpcCodes(codes []codes.Code) *Config {
	c.GrpcCodes = &codes
	return c
}
