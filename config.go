package relay

type State int

const (
	Closed State = iota
	HalfOpen
	Open
)

type Config struct {
	Name *string
	// time in seconds to move from open to half-open
	CoolDown *uint32
	// number of successful requsts to move  from half-open to closed
	SuccessesThreshold *uint32
	// number of failed requsts to move from closed to open
	FailuresThreshold *uint32
	// number of requsts allowed in half-open state
	HalfOpenRequestsQuota *uint32
	OnStateChange         *func(name string, from State, to State)
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
	c.OnStateChange = &onStateChange
	return c
}
