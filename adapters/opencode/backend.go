package opencode

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/jmeiracorbal/pinkman"
)

type Config struct {
	ID           string
	Executable   string
	Directory    string
	Capabilities []pinkman.Capability
	Priority     int
}

type Backend struct {
	config Config
}

func New(config Config) (*Backend, error) {
	if config.ID == "" {
		return nil, fmt.Errorf("opencode adapter: id is required")
	}
	if config.Executable == "" {
		config.Executable = "opencode"
	}
	return &Backend{config: config}, nil
}

func (b *Backend) ID() string { return b.config.ID }

func (b *Backend) Discover(ctx context.Context) ([]pinkman.Target, error) {
	path, err := exec.LookPath(b.config.Executable)
	if err != nil {
		return nil, fmt.Errorf("opencode adapter: executable %q not found: %w", b.config.Executable, err)
	}
	cmd := exec.CommandContext(ctx, path, "models")
	cmd.Dir = b.config.Directory
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("opencode adapter: list models: %w", err)
	}
	models := ParseModels(string(output))
	targets := make([]pinkman.Target, 0, len(models))
	for _, model := range models {
		targets = append(targets, pinkman.Target{
			Backend: b.config.ID, Model: model, Transport: pinkman.TransportCommand,
			Capabilities: append([]pinkman.Capability(nil), b.config.Capabilities...), Priority: b.config.Priority,
		})
	}
	return targets, nil
}

func (b *Backend) Execute(ctx context.Context, invocation pinkman.Invocation) (pinkman.Result, error) {
	path, err := exec.LookPath(b.config.Executable)
	if err != nil {
		return pinkman.Result{}, fmt.Errorf("opencode adapter: executable %q not found: %w", b.config.Executable, err)
	}
	args := []string{"run", "--standalone", "--format", "json"}
	if invocation.Target.Model != "" {
		args = append(args, "--model", invocation.Target.Model)
	}
	args = append(args, invocation.Input)
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Dir = b.config.Directory
	output, err := cmd.CombinedOutput()
	if err != nil {
		return pinkman.Result{}, fmt.Errorf("opencode adapter: run: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return pinkman.Result{Output: string(output), Raw: output}, nil
}

func ParseModels(output string) []string {
	seen := map[string]struct{}{}
	var models []string
	for _, line := range strings.Split(output, "\n") {
		model := strings.TrimSpace(line)
		if model == "" {
			continue
		}
		if _, exists := seen[model]; exists {
			continue
		}
		seen[model] = struct{}{}
		models = append(models, model)
	}
	sort.Strings(models)
	return models
}
