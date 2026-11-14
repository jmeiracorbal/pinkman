package pinkman

import (
	"context"
	"fmt"
	"sort"
)

type Discovery struct {
	Targets []Target
	Errors  map[string]error
}

type Catalog struct {
	backends map[string]Backend
}

func NewCatalog(backends ...Backend) (*Catalog, error) {
	catalog := &Catalog{backends: make(map[string]Backend, len(backends))}
	for _, backend := range backends {
		if backend == nil {
			return nil, fmt.Errorf("pinkman: nil backend")
		}
		id := backend.ID()
		if id == "" {
			return nil, fmt.Errorf("pinkman: backend id is empty")
		}
		if _, exists := catalog.backends[id]; exists {
			return nil, fmt.Errorf("pinkman: duplicate backend %q", id)
		}
		catalog.backends[id] = backend
	}
	return catalog, nil
}

func (c *Catalog) Discover(ctx context.Context) Discovery {
	ids := make([]string, 0, len(c.backends))
	for id := range c.backends {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	discovery := Discovery{Errors: make(map[string]error)}
	for _, id := range ids {
		targets, err := c.backends[id].Discover(ctx)
		if err != nil {
			discovery.Errors[id] = err
			continue
		}
		discovery.Targets = append(discovery.Targets, targets...)
	}

	sort.SliceStable(discovery.Targets, func(i, j int) bool {
		if discovery.Targets[i].Backend != discovery.Targets[j].Backend {
			return discovery.Targets[i].Backend < discovery.Targets[j].Backend
		}
		return discovery.Targets[i].Model < discovery.Targets[j].Model
	})
	return discovery
}

func (c *Catalog) Execute(ctx context.Context, invocation Invocation) (Result, error) {
	backend, ok := c.backends[invocation.Target.Backend]
	if !ok {
		return Result{}, fmt.Errorf("pinkman: unknown backend %q", invocation.Target.Backend)
	}
	return backend.Execute(ctx, invocation)
}
