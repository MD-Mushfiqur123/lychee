package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MD-Mushfiqur123/lychee/api"
)

type endpointState struct {
	endpoint            api.ModelEndpoint
	active              int32
	healthy             int32 // 1 = healthy (default), 0 = unhealthy
	lastCheck           time.Time
	lastError           string
	consecutiveFailures int32
	circuitOpenUntil    time.Time
}

type routeState struct {
	route         api.ModelRoute
	endpoints     []*endpointState
	weightedIndex []*endpointState
	counter       uint64
}

// ModelRouter load-balances requests to virtual models across defined backend endpoints.
type ModelRouter struct {
	mu        sync.RWMutex
	routes    map[string]*routeState
	storePath string
}

// NewModelRouter creates and initializes a ModelRouter.
func NewModelRouter(storePath string) *ModelRouter {
	mr := &ModelRouter{
		routes:    make(map[string]*routeState),
		storePath: storePath,
	}
	_ = mr.load()
	return mr
}

func (mr *ModelRouter) load() error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	data, err := os.ReadFile(mr.storePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var routesList []api.ModelRoute
	if err := json.Unmarshal(data, &routesList); err != nil {
		return err
	}

	for _, r := range routesList {
		endpoints := make([]*endpointState, len(r.Endpoints))
		var weightedIndex []*endpointState
		for i, ep := range r.Endpoints {
			state := &endpointState{endpoint: ep, healthy: 1}
			endpoints[i] = state
			w := ep.Weight
			if w <= 0 {
				w = 1
			}
			for j := 0; j < w; j++ {
				weightedIndex = append(weightedIndex, state)
			}
		}
		mr.routes[r.Name] = &routeState{
			route:         r,
			endpoints:     endpoints,
			weightedIndex: weightedIndex,
		}
	}
	return nil
}

func (mr *ModelRouter) persist() error {
	var routesList []api.ModelRoute
	for _, state := range mr.routes {
		routesList = append(routesList, state.route)
	}

	data, err := json.MarshalIndent(routesList, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(mr.storePath), 0755); err != nil {
		return err
	}

	return os.WriteFile(mr.storePath, data, 0644)
}

// AddRoute registers a new virtual model route.
func (mr *ModelRouter) AddRoute(route api.ModelRoute) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if route.Name == "" {
		return errors.New("route name is required")
	}
	if len(route.Endpoints) == 0 {
		return errors.New("at least one endpoint is required")
	}
	if route.Strategy == "" {
		route.Strategy = "round_robin"
	}

	endpoints := make([]*endpointState, len(route.Endpoints))
	var weightedIndex []*endpointState
	for i, ep := range route.Endpoints {
		state := &endpointState{endpoint: ep, healthy: 1}
		endpoints[i] = state
		w := ep.Weight
		if w <= 0 {
			w = 1
		}
		for j := 0; j < w; j++ {
			weightedIndex = append(weightedIndex, state)
		}
	}

	mr.routes[route.Name] = &routeState{
		route:         route,
		endpoints:     endpoints,
		weightedIndex: weightedIndex,
	}

	return mr.persist()
}

// RemoveRoute deletes a virtual model route.
func (mr *ModelRouter) RemoveRoute(name string) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if _, ok := mr.routes[name]; !ok {
		return errors.New("route not found")
	}

	delete(mr.routes, name)
	return mr.persist()
}

// ListRoutes lists all registered virtual model routes.
func (mr *ModelRouter) ListRoutes() []api.ModelRoute {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	routesList := make([]api.ModelRoute, 0, len(mr.routes))
	for _, state := range mr.routes {
		routesList = append(routesList, state.route)
	}
	return routesList
}

// Resolve matches a virtual name to a target endpoint and returns a callback to release the load tracking lock.
func (mr *ModelRouter) Resolve(name string) (*api.ModelEndpoint, func(), error) {
	mr.mu.RLock()
	state, ok := mr.routes[name]
	mr.mu.RUnlock()

	if !ok {
		return nil, func() {}, nil
	}

	if len(state.endpoints) == 0 {
		return nil, func() {}, errors.New("no endpoints for route")
	}

	// Filter to healthy endpoints only
	var available []*endpointState
	for _, ep := range state.endpoints {
		if atomic.LoadInt32(&ep.healthy) == 0 && !ep.circuitOpenUntil.IsZero() && time.Now().After(ep.circuitOpenUntil) {
			// Half-open: allow request to test
			atomic.StoreInt32(&ep.healthy, 1)
			slog.Info("router: circuit breaker HALF-OPEN", "host", ep.endpoint.Host)
		}
		if atomic.LoadInt32(&ep.healthy) == 1 {
			available = append(available, ep)
		}
	}
	if len(available) == 0 {
		slog.Warn("router: all endpoints unhealthy, falling through to local", "route", name)
		return nil, func() {}, nil
	}

	var chosen *endpointState

	switch state.route.Strategy {
	case "random":
		idx := rand.Intn(len(available))
		chosen = available[idx]
	case "least_loaded":
		chosen = available[0]
		minActive := atomic.LoadInt32(&chosen.active)
		for _, ep := range available[1:] {
			act := atomic.LoadInt32(&ep.active)
			if act < minActive {
				minActive = act
				chosen = ep
			}
		}
	case "weighted_round_robin":
		var availableWeighted []*endpointState
		for _, ep := range available {
			w := ep.endpoint.Weight
			if w <= 0 {
				w = 1
			}
			for j := 0; j < w; j++ {
				availableWeighted = append(availableWeighted, ep)
			}
		}
		val := atomic.AddUint64(&state.counter, 1)
		idx := (val - 1) % uint64(len(availableWeighted))
		chosen = availableWeighted[idx]
	case "round_robin":
		fallthrough
	default:
		val := atomic.AddUint64(&state.counter, 1)
		idx := (val - 1) % uint64(len(available))
		chosen = available[idx]
	}

	atomic.AddInt32(&chosen.active, 1)

	release := func() {
		atomic.AddInt32(&chosen.active, -1)
	}

	return &chosen.endpoint, release, nil
}

func (mr *ModelRouter) StartHealthChecks(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			mr.mu.RLock()
			for _, state := range mr.routes {
				for _, ep := range state.endpoints {
					go mr.checkEndpoint(ep)
				}
			}
			mr.mu.RUnlock()
		}
	}
}

func (mr *ModelRouter) checkEndpoint(ep *endpointState) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(ep.endpoint.Host + "/")
	ep.lastCheck = time.Now()

	if err != nil {
		atomic.StoreInt32(&ep.healthy, 0)
		ep.lastError = err.Error()
		slog.Warn("router: endpoint health check failed",
			"host", ep.endpoint.Host, "error", err)
		return
	}
	resp.Body.Close()

	if resp.StatusCode >= 500 {
		atomic.StoreInt32(&ep.healthy, 0)
		ep.lastError = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return
	}

	atomic.StoreInt32(&ep.healthy, 1)
	ep.lastError = ""
}

func (mr *ModelRouter) GetRouteStatus(name string) (*api.RouteStatusResponse, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	state, ok := mr.routes[name]
	if !ok {
		return nil, errors.New("route not found")
	}

	resp := &api.RouteStatusResponse{
		Name:     state.route.Name,
		Strategy: state.route.Strategy,
	}
	for _, ep := range state.endpoints {
		status := api.ModelEndpointStatus{
			Host:           ep.endpoint.Host,
			Model:          ep.endpoint.Model,
			Healthy:        atomic.LoadInt32(&ep.healthy) == 1,
			ActiveRequests: int(atomic.LoadInt32(&ep.active)),
			LastError:      ep.lastError,
		}
		if !ep.lastCheck.IsZero() {
			status.LastCheck = ep.lastCheck.Format(time.RFC3339)
		}
		resp.Endpoints = append(resp.Endpoints, status)
	}
	return resp, nil
}

func (mr *ModelRouter) RecordFailure(routeName, host string) {
	mr.mu.RLock()
	state := mr.routes[routeName]
	mr.mu.RUnlock()
	if state == nil {
		return
	}

	for _, ep := range state.endpoints {
		if ep.endpoint.Host == host {
			failures := atomic.AddInt32(&ep.consecutiveFailures, 1)
			if failures >= 3 {
				ep.circuitOpenUntil = time.Now().Add(30 * time.Second)
				atomic.StoreInt32(&ep.healthy, 0)
				slog.Warn("router: circuit breaker OPEN", "host", host, "failures", failures)
			}
			return
		}
	}
}

func (mr *ModelRouter) RecordSuccess(routeName, host string) {
	mr.mu.RLock()
	state := mr.routes[routeName]
	mr.mu.RUnlock()
	if state == nil {
		return
	}

	for _, ep := range state.endpoints {
		if ep.endpoint.Host == host {
			atomic.StoreInt32(&ep.consecutiveFailures, 0)
			atomic.StoreInt32(&ep.healthy, 1)
			return
		}
	}
}
