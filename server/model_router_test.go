package server

import (
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MD-Mushfiqur123/lychee/api"
)

func TestModelRouter(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "routes.json")
	router := NewModelRouter(storePath)

	route := api.ModelRoute{
		Name: "fast",
		Endpoints: []api.ModelEndpoint{
			{Host: "http://192.168.1.10:11434", Model: "gemma3"},
			{Host: "http://192.168.1.20:11434", Model: "gemma3"},
			{Host: "http://192.168.1.30:11434", Model: "gemma3"},
		},
		Strategy: "round_robin",
	}

	t.Run("add and list routes", func(t *testing.T) {
		err := router.AddRoute(route)
		if err != nil {
			t.Fatalf("failed to add route: %v", err)
		}

		routes := router.ListRoutes()
		if len(routes) != 1 {
			t.Fatalf("expected 1 route, got %d", len(routes))
		}
		if routes[0].Name != "fast" {
			t.Errorf("expected route name 'fast', got %q", routes[0].Name)
		}
	})

	t.Run("persist and load routes", func(t *testing.T) {
		router2 := NewModelRouter(storePath)
		routes := router2.ListRoutes()
		if len(routes) != 1 {
			t.Fatalf("expected 1 route on load, got %d", len(routes))
		}
		if routes[0].Name != "fast" {
			t.Errorf("expected loaded route name 'fast', got %q", routes[0].Name)
		}
	})

	t.Run("round robin resolution", func(t *testing.T) {
		ep1, rel1, _ := router.Resolve("fast")
		ep2, rel2, _ := router.Resolve("fast")
		ep3, rel3, _ := router.Resolve("fast")
		ep4, rel4, _ := router.Resolve("fast")

		defer rel1()
		defer rel2()
		defer rel3()
		defer rel4()

		if ep1.Host != "http://192.168.1.10:11434" {
			t.Errorf("expected first call to resolve host 10, got %q", ep1.Host)
		}
		if ep2.Host != "http://192.168.1.20:11434" {
			t.Errorf("expected second call to resolve host 20, got %q", ep2.Host)
		}
		if ep3.Host != "http://192.168.1.30:11434" {
			t.Errorf("expected third call to resolve host 30, got %q", ep3.Host)
		}
		if ep4.Host != "http://192.168.1.10:11434" {
			t.Errorf("expected fourth call to wrap around to host 10, got %q", ep4.Host)
		}
	})

	t.Run("random strategy resolution", func(t *testing.T) {
		randRoute := api.ModelRoute{
			Name: "random-route",
			Endpoints: []api.ModelEndpoint{
				{Host: "host1", Model: "m"},
				{Host: "host2", Model: "m"},
			},
			Strategy: "random",
		}
		_ = router.AddRoute(randRoute)

		for i := 0; i < 20; i++ {
			ep, rel, err := router.Resolve("random-route")
			if err != nil {
				t.Fatalf("resolve failed: %v", err)
			}
			rel()
			if ep.Host != "host1" && ep.Host != "host2" {
				t.Errorf("unexpected resolved host: %q", ep.Host)
			}
		}
	})

	t.Run("least loaded strategy resolution", func(t *testing.T) {
		llRoute := api.ModelRoute{
			Name: "least-loaded-route",
			Endpoints: []api.ModelEndpoint{
				{Host: "host1", Model: "m"},
				{Host: "host2", Model: "m"},
			},
			Strategy: "least_loaded",
		}
		_ = router.AddRoute(llRoute)

		// Both are active 0, host1 (first) is returned.
		epA, relA, _ := router.Resolve("least-loaded-route")
		
		// host1 is active = 1, host2 is active = 0. Resolves to host2.
		epB, relB, _ := router.Resolve("least-loaded-route")
		
		if epA.Host != "host1" {
			t.Errorf("expected epA to be host1, got %q", epA.Host)
		}
		if epB.Host != "host2" {
			t.Errorf("expected epB to be host2, got %q", epB.Host)
		}

		relA()
		relB()
	})

	t.Run("resolve unknown name", func(t *testing.T) {
		ep, _, err := router.Resolve("unknown-model")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ep != nil {
			t.Errorf("expected nil endpoint for unknown name, got %v", ep)
		}
	})

	t.Run("concurrent resolve safety", func(t *testing.T) {
		var wg sync.WaitGroup
		concurrency := 50

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ep, rel, err := router.Resolve("fast")
				if err == nil && ep != nil {
					rel()
				}
			}()
		}
		wg.Wait()
	})

	t.Run("remove route", func(t *testing.T) {
		err := router.RemoveRoute("fast")
		if err != nil {
			t.Fatalf("failed to remove route: %v", err)
		}

		routes := router.ListRoutes()
		for _, r := range routes {
			if r.Name == "fast" {
				t.Error("expected route 'fast' to be removed")
			}
		}

		ep, _, _ := router.Resolve("fast")
		if ep != nil {
			t.Error("expected resolved endpoint to be nil after deletion")
		}
	})

	t.Run("weighted round robin strategy resolution", func(t *testing.T) {
		weightedRoute := api.ModelRoute{
			Name: "weighted-route",
			Endpoints: []api.ModelEndpoint{
				{Host: "hostA", Model: "m", Weight: 3},
				{Host: "hostB", Model: "m", Weight: 1},
			},
			Strategy: "weighted_round_robin",
		}
		_ = router.AddRoute(weightedRoute)

		counts := make(map[string]int)
		for i := 0; i < 4; i++ {
			ep, rel, err := router.Resolve("weighted-route")
			if err != nil {
				t.Fatalf("resolve failed: %v", err)
			}
			counts[ep.Host]++
			rel()
		}

		if counts["hostA"] != 3 {
			t.Errorf("expected hostA to be chosen 3 times, got %d", counts["hostA"])
		}
		if counts["hostB"] != 1 {
			t.Errorf("expected hostB to be chosen 1 time, got %d", counts["hostB"])
		}
	})

	t.Run("skip unhealthy endpoints", func(t *testing.T) {
		unhealthyRoute := api.ModelRoute{
			Name: "unhealthy-route",
			Endpoints: []api.ModelEndpoint{
				{Host: "host1", Model: "m"},
				{Host: "host2", Model: "m"},
			},
			Strategy: "round_robin",
		}
		_ = router.AddRoute(unhealthyRoute)

		router.mu.RLock()
		state := router.routes["unhealthy-route"]
		router.mu.RUnlock()
		atomic.StoreInt32(&state.endpoints[0].healthy, 0)

		for i := 0; i < 5; i++ {
			ep, rel, err := router.Resolve("unhealthy-route")
			if err != nil {
				t.Fatalf("resolve failed: %v", err)
			}
			if ep.Host != "host2" {
				t.Errorf("expected only host2, got %q", ep.Host)
			}
			rel()
		}
	})

	t.Run("all endpoints unhealthy returns nil", func(t *testing.T) {
		allUnhealthyRoute := api.ModelRoute{
			Name: "all-unhealthy-route",
			Endpoints: []api.ModelEndpoint{
				{Host: "host1", Model: "m"},
			},
		}
		_ = router.AddRoute(allUnhealthyRoute)

		router.mu.RLock()
		state := router.routes["all-unhealthy-route"]
		router.mu.RUnlock()
		atomic.StoreInt32(&state.endpoints[0].healthy, 0)

		ep, _, err := router.Resolve("all-unhealthy-route")
		if err != nil {
			t.Fatalf("resolve failed: %v", err)
		}
		if ep != nil {
			t.Errorf("expected nil endpoint for all unhealthy route, got %v", ep)
		}
	})

	t.Run("circuit breaker behavior", func(t *testing.T) {
		cbRoute := api.ModelRoute{
			Name: "cb-route",
			Endpoints: []api.ModelEndpoint{
				{Host: "host-cb-1", Model: "m"},
				{Host: "host-cb-2", Model: "m"},
			},
			Strategy: "round_robin",
		}
		_ = router.AddRoute(cbRoute)

		router.RecordFailure("cb-route", "host-cb-1")
		router.RecordFailure("cb-route", "host-cb-1")
		router.RecordFailure("cb-route", "host-cb-1")

		for i := 0; i < 5; i++ {
			ep, rel, err := router.Resolve("cb-route")
			if err != nil {
				t.Fatalf("resolve failed: %v", err)
			}
			if ep.Host != "host-cb-2" {
				t.Errorf("expected only host-cb-2 to be chosen, got %q", ep.Host)
			}
			rel()
		}

		router.RecordSuccess("cb-route", "host-cb-1")
		ep, rel, _ := router.Resolve("cb-route")
		if ep == nil {
			t.Errorf("expected non-nil endpoint after circuit reset")
		}
		rel()
	})

	t.Run("circuit breaker half open recovery", func(t *testing.T) {
		cbRoute := api.ModelRoute{
			Name: "cb-half-open",
			Endpoints: []api.ModelEndpoint{
				{Host: "host-cb-half-1", Model: "m"},
			},
		}
		_ = router.AddRoute(cbRoute)

		router.RecordFailure("cb-half-open", "host-cb-half-1")
		router.RecordFailure("cb-half-open", "host-cb-half-1")
		router.RecordFailure("cb-half-open", "host-cb-half-1")

		ep, _, _ := router.Resolve("cb-half-open")
		if ep != nil {
			t.Errorf("expected nil endpoint when circuit is open")
		}

		router.mu.RLock()
		state := router.routes["cb-half-open"]
		router.mu.RUnlock()
		state.endpoints[0].circuitOpenUntil = time.Now().Add(-1 * time.Second)

		ep, rel, _ := router.Resolve("cb-half-open")
		if ep == nil || ep.Host != "host-cb-half-1" {
			t.Errorf("expected host-cb-half-1 on half-open retry, got %v", ep)
		}
		rel()

		if atomic.LoadInt32(&state.endpoints[0].healthy) != 1 {
			t.Errorf("expected endpoint to be marked healthy on half-open resolution")
		}
	})
}
