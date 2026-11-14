package pinkman

import "context"

type Transport string

const (
	TransportOpenAI Transport = "openai"
	TransportCommand Transport = "command"
)

type Capability string

const (
	CapabilityText      Capability = "text"
	CapabilityTools     Capability = "tools"
	CapabilityCoding    Capability = "coding"
	CapabilityStreaming Capability = "streaming"
)

type Target struct {
	Backend      string
	Model        string
	Transport    Transport
	Capabilities []Capability
	Priority     int
}

type RouteRequest struct {
	Backend              string
	Model                string
	RequiredCapabilities []Capability
}

type Candidate struct {
	Target Target
	Score  int
}

type Invocation struct {
	Target Target
	Input  string
}

type Result struct {
	Output string
	Raw    []byte
}

type Backend interface {
	ID() string
	Discover(context.Context) ([]Target, error)
	Execute(context.Context, Invocation) (Result, error)
}
