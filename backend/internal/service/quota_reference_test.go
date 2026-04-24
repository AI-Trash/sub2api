package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
)

func TestConvertBalanceToReferenceTokens(t *testing.T) {
	require.Equal(t, int64(1000000), convertBalanceToReferenceTokens(2.5, 0.0000025))
	require.Zero(t, convertBalanceToReferenceTokens(0, 0.0000025))
	require.Zero(t, convertBalanceToReferenceTokens(1, 0))
}

func TestApplyUsageProgressReferenceTokens(t *testing.T) {
	progress := &UsageProgress{
		Utilization: 25,
		WindowStats: &WindowStats{
			Cost: 2.5,
		},
	}

	applyUsageProgressReferenceTokens(progress, 0.0000025)

	require.Equal(t, int64(1000000), progress.UsedTokens)
	require.Equal(t, int64(4000000), progress.TotalTokens)
	require.Equal(t, int64(3000000), progress.RemainingTokens)
}

func TestApplyGroupSummaryReferenceTokens(t *testing.T) {
	summary := &usagestats.GroupUsageSummary{
		FiveHourBalance: 1.25,
		WeeklyBalance:   2.5,
		FiveHourTokens:  9,
		WeeklyTokens:    9,
	}

	applyGroupSummaryReferenceTokens(summary, 0.0000025)

	require.Equal(t, int64(500000), summary.FiveHourTokens)
	require.Equal(t, int64(1000000), summary.WeeklyTokens)
}

func TestGetReferenceTokenPrice_UsesNormalRatesOnly(t *testing.T) {
	svc := NewBillingService(&config.Config{}, nil)

	price, err := svc.GetReferenceTokenPrice("claude-sonnet-4")

	require.NoError(t, err)
	require.Equal(t, 9e-6, price)
}
