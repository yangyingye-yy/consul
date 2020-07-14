package usagemetrics

import (
	"time"

	"github.com/armon/go-metrics"
	"github.com/hashicorp/consul/agent/consul/state"
	"github.com/hashicorp/consul/agent/structs"
	"github.com/hashicorp/consul/logging"
	"github.com/hashicorp/go-hclog"
)

// Option type allows setting various optional parameters for the
// UsageMetricsReporter
type Option func(*UsageMetricsReporter)

// WithDatacenter adds the datacenter as a label to all metrics emitted by the
// UsageMetricsReporter
func WithDatacenter(dc string) Option {
	return func(u *UsageMetricsReporter) {
		u.metricLabels = append(u.metricLabels, metrics.Label{Name: "datacenter", Value: dc})
	}
}

// WithLogger takes a logger and creates a new, named sub-logger to use when
// running
func WithLogger(logger hclog.Logger) Option {
	return func(u *UsageMetricsReporter) {
		u.logger = logger.Named(logging.UsageMetrics)
	}
}

// WithReportingInterval specifies the interval on which UsageMetricsReporter
// should emit metrics
func WithReportingInterval(dur time.Duration) Option {
	return func(u *UsageMetricsReporter) {
		u.tickerInterval = dur
	}
}

// StateProvider defines an inteface for retrieving a state.Store handle. In
// non-test code, this is satisfied by the fsm.FSM struct.
type StateProvider interface {
	State() *state.Store
}

// UsageMetricsReporter provides functionality for emitting usage metrics into
// the metrics stream. This makes it essentially a translation layer
// between the state store and metrics stream.
type UsageMetricsReporter struct {
	logger         hclog.Logger
	metricLabels   []metrics.Label
	stateProvider  StateProvider
	tickerInterval time.Duration
}

func NewUsageMetricsReporter(sp StateProvider, opts ...Option) *UsageMetricsReporter {
	u := &UsageMetricsReporter{
		stateProvider: sp,
	}
	for _, o := range opts {
		o(u)
	}

	if u.logger == nil {
		u.logger = hclog.NewNullLogger()
	}

	if u.tickerInterval == 0 {
		// TODO (cpiraino): Pick a more well-informed default
		u.tickerInterval = 1 * time.Minute
	}
	return u
}

// Run must be run in a goroutine, and can be stopped by closing or sending
// data to the passed in shutdownCh
func (u *UsageMetricsReporter) Run(shutdownCh <-chan struct{}) {
	ticker := time.NewTicker(u.tickerInterval)
	for {
		select {
		case <-shutdownCh:
			u.logger.Debug("usage metrics reporter shutting down")
			ticker.Stop()
			return
		case <-ticker.C:
			u.runOnce()
		}
	}
}

func (u *UsageMetricsReporter) runOnce() {
	state := u.stateProvider.State()
	_, nodes, err := state.Nodes(nil)
	if err != nil {
		u.logger.Warn("failed to retrieve nodes from state store", "error", err)
	}
	metrics.SetGaugeWithLabels(
		[]string{"consul", "state", "nodes"},
		float32(len(nodes)),
		u.metricLabels,
	)

	_, services, err := state.ServiceList(nil, structs.WildcardEnterpriseMeta())
	if err != nil {
		u.logger.Warn("failed to retrieve services from state store", "error", err)
	}

	namespaceMap := make(map[string]int)
	// TODO (cpiraino): Deal with non-default namespaces
	namespaceMap[structs.DefaultEnterpriseMeta().NamespaceOrDefault()] = 0
	for _, svc := range services {
		namespaceMap[svc.NamespaceOrDefault()] += 1
	}

	for ns, count := range namespaceMap {
		metrics.SetGaugeWithLabels(
			[]string{"consul", "state", "services"},
			float32(count),
			append(u.metricLabels, metrics.Label{Name: "namespace", Value: ns}),
		)
	}
}
