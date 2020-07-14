package usagemetrics

import (
	"testing"
	"time"

	"github.com/armon/go-metrics"
	"github.com/hashicorp/consul/agent/consul/state"
	"github.com/hashicorp/consul/agent/structs"
	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockStateProvider struct {
	mock.Mock
}

func (m *mockStateProvider) State() *state.Store {
	retValues := m.Called()
	return retValues.Get(0).(*state.Store)
}

func TestUsageReporter_Run(t *testing.T) {
	type testCase struct {
		modfiyStateStore func(t *testing.T, s *state.Store)
		expectedGauges   map[string]metrics.GaugeValue
	}
	cases := map[string]testCase{
		"empty-state": testCase{
			expectedGauges: map[string]metrics.GaugeValue{
				"consul.usage.test.consul.state.nodes;datacenter=dc1": metrics.GaugeValue{
					Name:   "consul.usage.test.consul.state.nodes",
					Value:  0,
					Labels: []metrics.Label{{Name: "datacenter", Value: "dc1"}},
				},
				"consul.usage.test.consul.state.services;datacenter=dc1;namespace=default": metrics.GaugeValue{
					Name:  "consul.usage.test.consul.state.services",
					Value: 0,
					Labels: []metrics.Label{
						{Name: "datacenter", Value: "dc1"},
						{Name: "namespace", Value: "default"},
					},
				},
			},
		},
		"nodes-and-services": testCase{
			modfiyStateStore: func(t *testing.T, s *state.Store) {
				require.Nil(t, s.EnsureNode(1, &structs.Node{Node: "foo", Address: "127.0.0.1"}))
				require.Nil(t, s.EnsureNode(2, &structs.Node{Node: "bar", Address: "127.0.0.2"}))
				require.Nil(t, s.EnsureNode(3, &structs.Node{Node: "baz", Address: "127.0.0.2"}))

				// Typical services and some consul services spread across two nodes
				require.Nil(t, s.EnsureService(4, "foo", &structs.NodeService{ID: "db", Service: "db", Tags: nil, Address: "", Port: 5000}))
				require.Nil(t, s.EnsureService(5, "bar", &structs.NodeService{ID: "api", Service: "api", Tags: nil, Address: "", Port: 5000}))
				require.Nil(t, s.EnsureService(6, "bar", &structs.NodeService{ID: "consul", Service: "consul", Tags: nil}))
				require.Nil(t, s.EnsureService(7, "bar", &structs.NodeService{ID: "consul", Service: "consul", Tags: nil}))
			},
			expectedGauges: map[string]metrics.GaugeValue{
				"consul.usage.test.consul.state.nodes;datacenter=dc1": metrics.GaugeValue{
					Name:   "consul.usage.test.consul.state.nodes",
					Value:  3,
					Labels: []metrics.Label{{Name: "datacenter", Value: "dc1"}},
				},
				"consul.usage.test.consul.state.services;datacenter=dc1;namespace=default": metrics.GaugeValue{
					Name:  "consul.usage.test.consul.state.services",
					Value: 3,
					Labels: []metrics.Label{
						{Name: "datacenter", Value: "dc1"},
						{Name: "namespace", Value: "default"},
					},
				},
			},
		},
	}

	for name, tcase := range cases {
		t.Run(name, func(t *testing.T) {
			// Only have a single interval for the test
			sink := metrics.NewInmemSink(1*time.Minute, 1*time.Minute)
			cfg := metrics.DefaultConfig("consul.usage.test")
			cfg.EnableHostname = false
			metrics.NewGlobal(cfg, sink)

			mockStateProvider := &mockStateProvider{}
			s, err := state.NewStateStore(nil)
			require.NoError(t, err)
			if tcase.modfiyStateStore != nil {
				tcase.modfiyStateStore(t, s)
			}
			mockStateProvider.On("State").Return(s)

			reporter := NewUsageMetricsReporter(
				mockStateProvider,
				WithLogger(testutil.Logger(t)),
				WithDatacenter("dc1"),
			)
			require.NoError(t, err)

			reporter.runOnce()

			intervals := sink.Data()
			require.Len(t, intervals, 1)
			intv := intervals[0]
			require.Equal(t, tcase.expectedGauges, intv.Gauges)
		})
	}
}
