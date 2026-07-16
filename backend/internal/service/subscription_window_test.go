//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type subscriptionWindowRepoStub struct {
	UserSubscriptionRepository

	activatedStart *time.Time
	dailyStart     *time.Time
	weeklyStart    *time.Time
	monthlyStart   *time.Time
}

func (r *subscriptionWindowRepoStub) ActivateWindows(_ context.Context, _ int64, start time.Time) error {
	r.activatedStart = &start
	return nil
}

func (r *subscriptionWindowRepoStub) ResetDailyUsage(_ context.Context, _ int64, _ *time.Time, start time.Time) error {
	r.dailyStart = &start
	return nil
}

func (r *subscriptionWindowRepoStub) ResetWeeklyUsage(_ context.Context, _ int64, _ *time.Time, start time.Time) error {
	r.weeklyStart = &start
	return nil
}

func (r *subscriptionWindowRepoStub) ResetMonthlyUsage(_ context.Context, _ int64, _ *time.Time, start time.Time) error {
	r.monthlyStart = &start
	return nil
}

func TestSubscriptionWindows_DoNotResetBeforeSubscriptionAnchoredWindowExpires(t *testing.T) {
	loc := time.FixedZone("CST", 8*60*60)
	startsAt := time.Date(2026, 5, 1, 18, 0, 0, 0, loc)

	legacyMidnightDaily := time.Date(2026, 5, 1, 0, 0, 0, 0, loc)
	legacyMidnightWeekly := time.Date(2026, 4, 25, 0, 0, 0, 0, loc)
	legacyMidnightMonthly := time.Date(2026, 4, 2, 0, 0, 0, 0, loc)

	tests := []struct {
		name        string
		windowStart *time.Time
		duration    time.Duration
		now         time.Time
		wantReset   bool
	}{
		{
			name:        "daily subscription bought at 18:00 does not reset at next midnight",
			windowStart: &legacyMidnightDaily,
			duration:    subscriptionDailyWindow,
			now:         time.Date(2026, 5, 2, 0, 1, 0, 0, loc),
			wantReset:   false,
		},
		{
			name:        "weekly subscription does not reset at calendar week boundary",
			windowStart: &legacyMidnightWeekly,
			duration:    subscriptionWeeklyWindow,
			now:         time.Date(2026, 5, 8, 0, 1, 0, 0, loc),
			wantReset:   false,
		},
		{
			name:        "monthly subscription does not reset before 30 day rolling boundary",
			windowStart: &legacyMidnightMonthly,
			duration:    subscriptionMonthlyWindow,
			now:         time.Date(2026, 5, 31, 0, 1, 0, 0, loc),
			wantReset:   false,
		},
		{
			name:        "daily resets once 24 hours from subscription start have elapsed",
			windowStart: &legacyMidnightDaily,
			duration:    subscriptionDailyWindow,
			now:         time.Date(2026, 5, 2, 18, 0, 0, 0, loc),
			wantReset:   true,
		},
		{
			name:        "weekly resets once 7 days from subscription start have elapsed",
			windowStart: &legacyMidnightWeekly,
			duration:    subscriptionWeeklyWindow,
			now:         time.Date(2026, 5, 8, 18, 0, 0, 0, loc),
			wantReset:   true,
		},
		{
			name:        "monthly resets once 30 days from subscription start have elapsed",
			windowStart: &legacyMidnightMonthly,
			duration:    subscriptionMonthlyWindow,
			now:         time.Date(2026, 5, 31, 18, 0, 0, 0, loc),
			wantReset:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &UserSubscription{StartsAt: startsAt}
			got := sub.needsWindowResetAt(tt.windowStart, tt.duration, tt.now)
			require.Equal(t, tt.wantReset, got)
		})
	}
}

func TestSubscriptionResetWindowStart_RollsFromSubscriptionStart(t *testing.T) {
	loc := time.FixedZone("CST", 8*60*60)
	startsAt := time.Date(2026, 5, 1, 18, 0, 0, 0, loc)
	legacyMidnight := time.Date(2026, 5, 1, 0, 0, 0, 0, loc)

	sub := &UserSubscription{StartsAt: startsAt}
	now := time.Date(2026, 5, 3, 19, 0, 0, 0, loc)

	got := subscriptionResetWindowStart(sub, &legacyMidnight, subscriptionDailyWindow, now)

	require.Equal(t, time.Date(2026, 5, 3, 18, 0, 0, 0, loc), got)
}

func TestCheckAndActivateWindow_UsesSubscriptionStartAsAnchor(t *testing.T) {
	loc := time.FixedZone("CST", 8*60*60)
	startsAt := time.Date(2026, 5, 1, 18, 0, 0, 0, loc)
	repo := &subscriptionWindowRepoStub{}
	svc := NewSubscriptionService(nil, repo, nil, nil, nil)

	err := svc.CheckAndActivateWindow(context.Background(), &UserSubscription{
		ID:        1,
		StartsAt:  startsAt,
		ExpiresAt: startsAt.Add(24 * time.Hour),
		Status:    SubscriptionStatusActive,
	})

	require.NoError(t, err)
	require.NotNil(t, repo.activatedStart)
	require.Equal(t, startsAt, *repo.activatedStart)
}

func TestNormalizeExpiredWindows_ReturnsEffectiveRollingWindowStart(t *testing.T) {
	loc := time.FixedZone("CST", 8*60*60)
	now := time.Now().In(loc)
	startsAt := now.Add(-2*subscriptionDailyWindow - time.Hour).Truncate(time.Second)
	legacyMidnight := time.Date(startsAt.Year(), startsAt.Month(), startsAt.Day(), 0, 0, 0, 0, loc)

	subs := []UserSubscription{
		{
			StartsAt:           startsAt,
			ExpiresAt:          now.Add(subscriptionDailyWindow),
			Status:             SubscriptionStatusActive,
			DailyWindowStart:   &legacyMidnight,
			WeeklyWindowStart:  &legacyMidnight,
			MonthlyWindowStart: &legacyMidnight,
			DailyUsageUSD:      1,
			WeeklyUsageUSD:     2,
			MonthlyUsageUSD:    3,
		},
	}

	normalizeExpiredWindows(subs)

	require.NotNil(t, subs[0].DailyWindowStart)
	require.Equal(t, startsAt.Add(2*subscriptionDailyWindow), *subs[0].DailyWindowStart)
	require.NotNil(t, subs[0].WeeklyWindowStart)
	require.Equal(t, startsAt, *subs[0].WeeklyWindowStart)
	require.NotNil(t, subs[0].MonthlyWindowStart)
	require.Equal(t, startsAt, *subs[0].MonthlyWindowStart)
}

func TestCalculateProgress_UsesEffectiveRollingWindowStart(t *testing.T) {
	svc := newTestSubscriptionService()
	loc := time.FixedZone("CST", 8*60*60)
	now := time.Now().In(loc)
	startsAt := now.Add(-2*subscriptionDailyWindow - time.Hour).Truncate(time.Second)
	legacyMidnight := time.Date(startsAt.Year(), startsAt.Month(), startsAt.Day(), 0, 0, 0, 0, loc)

	sub := &UserSubscription{
		ID:               1,
		StartsAt:         startsAt,
		ExpiresAt:        now.Add(subscriptionDailyWindow),
		DailyUsageUSD:    3,
		DailyWindowStart: &legacyMidnight,
	}
	group := &Group{
		Name:          "Pro",
		DailyLimitUSD: ptrFloat64(10),
	}

	progress := svc.calculateProgress(sub, group)

	require.NotNil(t, progress.Daily)
	require.Equal(t, startsAt.Add(2*subscriptionDailyWindow), progress.Daily.WindowStart)
	require.Equal(t, 0.0, progress.Daily.UsedUSD)
}
