package pinkman

import (
	"errors"
	"reflect"
	"testing"
)

func TestResolverFiltersAndRanksDeterministically(t *testing.T) {
	targets := []Target{
		{Backend: "local", Model: "small", Capabilities: []Capability{CapabilityText}, Priority: 100},
		{Backend: "opencode", Model: "coder-b", Capabilities: []Capability{CapabilityText, CapabilityCoding}, Priority: 10},
		{Backend: "opencode", Model: "coder-a", Capabilities: []Capability{CapabilityText, CapabilityCoding}, Priority: 10},
	}
	got, err := (Resolver{}).Resolve(RouteRequest{RequiredCapabilities: []Capability{CapabilityCoding}}, targets)
	if err != nil {
		t.Fatal(err)
	}
	models := []string{got[0].Target.Model, got[1].Target.Model}
	want := []string{"coder-a", "coder-b"}
	if !reflect.DeepEqual(models, want) {
		t.Fatalf("models = %v, want %v", models, want)
	}
}

func TestResolverHonorsExplicitModel(t *testing.T) {
	targets := []Target{{Backend: "opencode", Model: "a"}, {Backend: "opencode", Model: "b"}}
	got, err := (Resolver{}).Resolve(RouteRequest{Backend: "opencode", Model: "b"}, targets)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Target.Model != "b" {
		t.Fatalf("unexpected candidates: %#v", got)
	}
}

func TestResolverReturnsTypedError(t *testing.T) {
	_, err := (Resolver{}).Resolve(RouteRequest{Model: "missing"}, []Target{{Model: "present"}})
	if !errors.Is(err, ErrNoCompatibleTarget) {
		t.Fatalf("err = %v, want %v", err, ErrNoCompatibleTarget)
	}
}
