package pinkman

import (
	"errors"
	"sort"
)

var ErrNoCompatibleTarget = errors.New("pinkman: no compatible target")

type Resolver struct{}

func (Resolver) Resolve(req RouteRequest, targets []Target) ([]Candidate, error) {
	candidates := make([]Candidate, 0, len(targets))
	for _, target := range targets {
		if req.Backend != "" && target.Backend != req.Backend {
			continue
		}
		if req.Model != "" && target.Model != req.Model {
			continue
		}
		if !hasAll(target.Capabilities, req.RequiredCapabilities) {
			continue
		}

		score := target.Priority
		if req.Backend != "" && target.Backend == req.Backend {
			score += 100
		}
		if req.Model != "" && target.Model == req.Model {
			score += 1000
		}
		candidates = append(candidates, Candidate{Target: target, Score: score})
	}

	if len(candidates) == 0 {
		return nil, ErrNoCompatibleTarget
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		if candidates[i].Target.Backend != candidates[j].Target.Backend {
			return candidates[i].Target.Backend < candidates[j].Target.Backend
		}
		return candidates[i].Target.Model < candidates[j].Target.Model
	})
	return candidates, nil
}

func hasAll(have, required []Capability) bool {
	set := make(map[Capability]struct{}, len(have))
	for _, capability := range have {
		set[capability] = struct{}{}
	}
	for _, capability := range required {
		if _, ok := set[capability]; !ok {
			return false
		}
	}
	return true
}
