package metrics_test

import (
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/relixdev/relix/relay/internal/metrics"
)

func TestActiveConnectionsRegistered(t *testing.T) {
	// Increment and verify the gauge is functional.
	metrics.ActiveConnections.WithLabelValues("agent").Add(2)
	metrics.ActiveConnections.WithLabelValues("mobile").Add(3)

	var m dto.Metric

	if err := metrics.ActiveConnections.WithLabelValues("agent").Write(&m); err != nil {
		t.Fatalf("write agent gauge: %v", err)
	}
	if got := m.GetGauge().GetValue(); got != 2 {
		t.Errorf("agent connections: want 2, got %v", got)
	}

	if err := metrics.ActiveConnections.WithLabelValues("mobile").Write(&m); err != nil {
		t.Fatalf("write mobile gauge: %v", err)
	}
	if got := m.GetGauge().GetValue(); got != 3 {
		t.Errorf("mobile connections: want 3, got %v", got)
	}

	// Reset so tests are independent.
	metrics.ActiveConnections.Reset()
}

func TestMessagesRoutedTotalIncrement(t *testing.T) {
	metrics.MessagesRoutedTotal.WithLabelValues("agent_to_mobile").Inc()
	metrics.MessagesRoutedTotal.WithLabelValues("agent_to_mobile").Inc()
	metrics.MessagesRoutedTotal.WithLabelValues("mobile_to_agent").Inc()

	var m dto.Metric

	if err := metrics.MessagesRoutedTotal.WithLabelValues("agent_to_mobile").Write(&m); err != nil {
		t.Fatalf("write counter: %v", err)
	}
	if got := m.GetCounter().GetValue(); got != 2 {
		t.Errorf("agent_to_mobile: want 2, got %v", got)
	}

	if err := metrics.MessagesRoutedTotal.WithLabelValues("mobile_to_agent").Write(&m); err != nil {
		t.Fatalf("write counter: %v", err)
	}
	if got := m.GetCounter().GetValue(); got != 1 {
		t.Errorf("mobile_to_agent: want 1, got %v", got)
	}

	metrics.MessagesRoutedTotal.Reset()
}

func TestBufferSizeGauge(t *testing.T) {
	metrics.BufferSize.WithLabelValues("user42").Set(5)

	var m dto.Metric
	if err := metrics.BufferSize.WithLabelValues("user42").Write(&m); err != nil {
		t.Fatalf("write gauge: %v", err)
	}
	if got := m.GetGauge().GetValue(); got != 5 {
		t.Errorf("buffer size: want 5, got %v", got)
	}

	metrics.BufferSize.Reset()
}
