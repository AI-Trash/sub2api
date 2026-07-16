package service

import "time"

const (
	subscriptionDailyWindow   = 24 * time.Hour
	subscriptionWeeklyWindow  = 7 * subscriptionDailyWindow
	subscriptionMonthlyWindow = 30 * subscriptionDailyWindow
)

type UserSubscription struct {
	ID      int64
	UserID  int64
	GroupID int64

	StartsAt  time.Time
	ExpiresAt time.Time
	Status    string

	DailyWindowStart   *time.Time
	WeeklyWindowStart  *time.Time
	MonthlyWindowStart *time.Time

	DailyUsageUSD   float64
	WeeklyUsageUSD  float64
	MonthlyUsageUSD float64

	AssignedBy *int64
	AssignedAt time.Time
	Notes      string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	User           *User
	Group          *Group
	AssignedByUser *User
}

func (s *UserSubscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive && time.Now().Before(s.ExpiresAt)
}

func (s *UserSubscription) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *UserSubscription) DaysRemaining() int {
	if s.IsExpired() {
		return 0
	}
	return int(time.Until(s.ExpiresAt).Hours() / 24)
}

func (s *UserSubscription) IsWindowActivated() bool {
	return s.DailyWindowStart != nil || s.WeeklyWindowStart != nil || s.MonthlyWindowStart != nil
}

func (s *UserSubscription) HasOneTimeDailyQuota() bool {
	if s == nil || s.StartsAt.IsZero() || s.ExpiresAt.IsZero() {
		return false
	}
	return !s.ExpiresAt.After(s.StartsAt.AddDate(0, 0, 1))
}

func (s *UserSubscription) NeedsDailyReset() bool {
	return s.NeedsDailyResetAt(time.Now())
}

func (s *UserSubscription) NeedsDailyResetAt(now time.Time) bool {
	if s.DailyWindowStart == nil {
		return false
	}
	if s.HasOneTimeDailyQuota() {
		return false
	}
	return !now.Before(s.DailyWindowStart.Add(subscriptionDailyWindow))
}

func (s *UserSubscription) NeedsWeeklyReset() bool {
	return s.needsWindowResetAt(s.WeeklyWindowStart, subscriptionWeeklyWindow, time.Now())
}

func (s *UserSubscription) NeedsMonthlyReset() bool {
	return s.needsWindowResetAt(s.MonthlyWindowStart, subscriptionMonthlyWindow, time.Now())
}

func (s *UserSubscription) DailyResetTime() *time.Time {
	if s.DailyWindowStart == nil {
		return nil
	}
	if s.HasOneTimeDailyQuota() {
		t := s.ExpiresAt
		return &t
	}
	return s.windowResetTimeAt(s.DailyWindowStart, subscriptionDailyWindow, time.Now())
}

func (s *UserSubscription) WeeklyResetTime() *time.Time {
	return s.windowResetTimeAt(s.WeeklyWindowStart, subscriptionWeeklyWindow, time.Now())
}

func (s *UserSubscription) MonthlyResetTime() *time.Time {
	return s.windowResetTimeAt(s.MonthlyWindowStart, subscriptionMonthlyWindow, time.Now())
}

func (s *UserSubscription) needsWindowResetAt(windowStart *time.Time, duration time.Duration, now time.Time) bool {
	start, ok := s.effectiveWindowStart(windowStart)
	if !ok {
		return false
	}
	return !now.Before(start.Add(duration))
}

func (s *UserSubscription) currentWindowStartAt(windowStart *time.Time, duration time.Duration, now time.Time) (time.Time, bool) {
	start, ok := s.effectiveWindowStart(windowStart)
	if !ok {
		return time.Time{}, false
	}
	return subscriptionWindowStartForNow(start, duration, now), true
}

func (s *UserSubscription) windowResetTimeAt(windowStart *time.Time, duration time.Duration, now time.Time) *time.Time {
	start, ok := s.currentWindowStartAt(windowStart, duration, now)
	if !ok {
		return nil
	}
	t := start.Add(duration)
	return &t
}

func (s *UserSubscription) effectiveWindowStart(windowStart *time.Time) (time.Time, bool) {
	if windowStart == nil {
		return time.Time{}, false
	}
	start := *windowStart
	if !s.StartsAt.IsZero() && start.Before(s.StartsAt) {
		start = s.StartsAt
	}
	return start, true
}

func subscriptionWindowStartForNow(anchor time.Time, duration time.Duration, now time.Time) time.Time {
	if duration <= 0 || anchor.IsZero() || now.Before(anchor) {
		return anchor
	}
	periods := int64(now.Sub(anchor) / duration)
	return anchor.Add(time.Duration(periods) * duration)
}

func (s *UserSubscription) CheckDailyLimit(group *Group, additionalCost float64) bool {
	if !group.HasDailyLimit() {
		return true
	}
	return s.DailyUsageUSD+additionalCost <= *group.DailyLimitUSD
}

func (s *UserSubscription) CheckWeeklyLimit(group *Group, additionalCost float64) bool {
	if !group.HasWeeklyLimit() {
		return true
	}
	return s.WeeklyUsageUSD+additionalCost <= *group.WeeklyLimitUSD
}

func (s *UserSubscription) CheckMonthlyLimit(group *Group, additionalCost float64) bool {
	if !group.HasMonthlyLimit() {
		return true
	}
	return s.MonthlyUsageUSD+additionalCost <= *group.MonthlyLimitUSD
}

func (s *UserSubscription) CheckAllLimits(group *Group, additionalCost float64) (daily, weekly, monthly bool) {
	daily = s.CheckDailyLimit(group, additionalCost)
	weekly = s.CheckWeeklyLimit(group, additionalCost)
	monthly = s.CheckMonthlyLimit(group, additionalCost)
	return
}
