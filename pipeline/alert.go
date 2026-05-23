package pipeline

import (
	"fmt"
	"sync"
	"time"
)

// AlertRule defines a condition that triggers an alert.
type AlertRule struct {
	Name      string
	Level     string
	Contains  string
	Threshold int
	Window    time.Duration
	OnAlert   func(rule string, count int, sample LogEntry)
}

// Alerter monitors log entries against alert rules.
type Alerter struct {
	mu      sync.Mutex
	rules   []AlertRule
	counts  map[string][]time.Time
}

// NewAlerter creates an Alerter with the given rules.
func NewAlerter(rules []AlertRule) *Alerter {
	return &Alerter{
		rules:  rules,
		counts: make(map[string][]time.Time),
	}
}

// Evaluate checks a log entry against all alert rules.
func (a *Alerter) Evaluate(entry LogEntry) {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	for _, rule := range a.rules {
		if !a.matches(rule, entry) {
			continue
		}
		key := rule.Name
		a.counts[key] = append(a.counts[key], now)
		a.counts[key] = a.prune(a.counts[key], now, rule.Window)
		if len(a.counts[key]) >= rule.Threshold {
			if rule.OnAlert != nil {
				rule.OnAlert(rule.Name, len(a.counts[key]), entry)
			}
			a.counts[key] = nil
		}
	}
}

func (a *Alerter) matches(rule AlertRule, entry LogEntry) bool {
	if rule.Level != "" && normalizeLevel(entry.Level) != normalizeLevel(rule.Level) {
		return false
	}
	if rule.Contains != "" && !containsString(entry.Message, rule.Contains) {
		return false
	}
	return true
}

func (a *Alerter) prune(times []time.Time, now time.Time, window time.Duration) []time.Time {
	cutoff := now.Add(-window)
	var result []time.Time
	for _, t := range times {
		if t.After(cutoff) {
			result = append(result, t)
		}
	}
	return result
}

// AlertRuleFromConfig builds an AlertRule from config-style map values.
func AlertRuleFromConfig(name, level, contains string, threshold int, window time.Duration, onAlert func(string, int, LogEntry)) AlertRule {
	if threshold <= 0 {
		threshold = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return AlertRule{
		Name:      fmt.Sprintf("%s", name),
		Level:     level,
		Contains:  contains,
		Threshold: threshold,
		Window:    window,
		OnAlert:   onAlert,
	}
}
