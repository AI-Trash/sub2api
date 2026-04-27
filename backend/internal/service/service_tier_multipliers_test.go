package service

import (
	"math"
	"testing"
)

func TestConfiguredServiceTierMultiplierOverridesPriorityPrices(t *testing.T) {
	svc := &BillingService{}
	pricing := &ModelPricing{
		InputPricePerToken:             1,
		InputPricePerTokenPriority:     100,
		OutputPricePerToken:            2,
		OutputPricePerTokenPriority:    200,
		CacheCreationPricePerToken:     3,
		CacheReadPricePerToken:         4,
		CacheReadPricePerTokenPriority: 400,
		ServiceTierMultipliers:         map[string]float64{"priority": 1.25},
	}
	tokens := UsageTokens{
		InputTokens:         10,
		OutputTokens:        10,
		CacheCreationTokens: 10,
		CacheReadTokens:     10,
	}

	cost := svc.computeTokenBreakdown(pricing, tokens, 1.0, "priority", false)

	assertFloatClose(t, 12.5, cost.InputCost)
	assertFloatClose(t, 25.0, cost.OutputCost)
	assertFloatClose(t, 37.5, cost.CacheCreationCost)
	assertFloatClose(t, 50.0, cost.CacheReadCost)
	assertFloatClose(t, 125.0, cost.TotalCost)
}

func TestConfiguredServiceTierMultiplierAllowsZero(t *testing.T) {
	svc := &BillingService{}
	pricing := &ModelPricing{
		InputPricePerToken:         1,
		OutputPricePerToken:        2,
		CacheCreationPricePerToken: 3,
		CacheReadPricePerToken:     4,
		ServiceTierMultipliers:     map[string]float64{"flex": 0},
	}

	cost := svc.computeTokenBreakdown(pricing, UsageTokens{
		InputTokens:         10,
		OutputTokens:        10,
		CacheCreationTokens: 10,
		CacheReadTokens:     10,
	}, 1.0, "flex", false)

	assertFloatClose(t, 0, cost.TotalCost)
	assertFloatClose(t, 0, cost.ActualCost)
}

func assertFloatClose(t *testing.T, want, got float64) {
	t.Helper()
	if math.Abs(want-got) > 1e-12 {
		t.Fatalf("expected %v, got %v", want, got)
	}
}
