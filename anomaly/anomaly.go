package anomaly

import (
	"fmt"
	"time"

	"github.com/mailgun/vulcand/Godeps/_workspace/src/github.com/mailgun/vulcan/metrics"
	"github.com/mailgun/vulcand/backend"
)

const (
	CodeLatency = iota + 1
	CodeNetErrorRate
	CodeAppErrorRate
)

const (
	MessageNetErrRate = "Error rate stands out"
	MessageAppErrRate = "App error rate (status 500) stands out"
	MessageLatency    = "%0.2f quantile latency stands out"
)

// MarkEndpointAnomalies takes the list of endpoints and marks anomalies detected within this set
// by modifying the inner Verdict property.
func MarkEndpointAnomalies(endpoints []*backend.Endpoint) {
	if len(endpoints) == 0 {
		return
	}

	stats := make([]*backend.RoundTripStats, len(endpoints))
	for i, e := range endpoints {
		stats[i] = &e.Stats
	}
	MarkAnomalies(stats)
}

// MarkAnomalies takes the list of stats and marks anomalies detected within this group by updating
// the Verdict property.
func MarkAnomalies(stats []*backend.RoundTripStats) {
	if len(stats) == 0 {
		return
	}
	markLatencies(stats)
	markNetErrorRates(stats)
	markAppErrorRates(stats)
}

func markNetErrorRates(stats []*backend.RoundTripStats) {
	errRates := make([]float64, len(stats))
	for i, s := range stats {
		errRates[i] = s.NetErrorRate()
	}

	_, bad := metrics.SplitRatios(errRates)
	for _, s := range stats {
		if bad[s.NetErrorRate()] {
			s.Verdict.IsBad = true
			s.Verdict.Anomalies = append(s.Verdict.Anomalies, backend.Anomaly{Code: CodeNetErrorRate, Message: MessageNetErrRate})
		}
	}
}

func markLatencies(stats []*backend.RoundTripStats) {
	// We are processing only median as others are more volatile
	markLatency(0, stats)
}

func markLatency(index int, stats []*backend.RoundTripStats) {
	quantiles := make([]time.Duration, len(stats))
	for i, s := range stats {
		quantiles[i] = s.LatencyBrackets[index].Value
	}

	quantile := stats[0].LatencyBrackets[index].Quantile
	_, bad := metrics.SplitLatencies(quantiles, time.Millisecond)
	for _, s := range stats {
		if bad[s.LatencyBrackets[index].Value] {
			s.Verdict.IsBad = true
			s.Verdict.Anomalies = append(
				s.Verdict.Anomalies,
				backend.Anomaly{
					Code:    CodeLatency,
					Message: fmt.Sprintf(MessageLatency, quantile),
				})
		}
	}
}

func markAppErrorRates(stats []*backend.RoundTripStats) {
	errRates := make([]float64, len(stats))
	for i, s := range stats {
		errRates[i] = s.AppErrorRate()
	}

	_, bad := metrics.SplitRatios(errRates)
	for _, s := range stats {
		if bad[s.AppErrorRate()] {
			s.Verdict.IsBad = true
			s.Verdict.Anomalies = append(
				s.Verdict.Anomalies, backend.Anomaly{Code: CodeAppErrorRate, Message: MessageAppErrRate})
		}
	}
}