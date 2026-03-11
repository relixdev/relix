package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ActiveConnections tracks the number of currently connected clients,
	// labelled by role (agent or mobile).
	ActiveConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "relay_active_connections",
			Help: "Number of currently active WebSocket connections.",
		},
		[]string{"role"},
	)

	// MessagesRoutedTotal counts every successfully routed envelope,
	// labelled by direction (agent_to_mobile or mobile_to_agent).
	MessagesRoutedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "relay_messages_routed_total",
			Help: "Total number of envelopes routed between agents and mobiles.",
		},
		[]string{"direction"},
	)

	// BufferSize tracks the total number of buffered messages waiting for
	// an offline mobile to reconnect, labelled by user_id.
	BufferSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "relay_buffer_size",
			Help: "Number of envelopes currently buffered per user.",
		},
		[]string{"user_id"},
	)

	// MessageLatency records the end-to-end routing latency in seconds.
	MessageLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "relay_message_latency_seconds",
			Help:    "End-to-end routing latency from envelope receipt to delivery.",
			Buckets: prometheus.DefBuckets,
		},
	)
)
