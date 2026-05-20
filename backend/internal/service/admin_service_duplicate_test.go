package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type duplicateAccountRepoStub struct {
	source         *Account
	created        *Account
	boundAccountID int64
	boundGroupIDs  []int64
}

func (s *duplicateAccountRepoStub) GetByID(ctx context.Context, id int64) (*Account, error) {
	if s.created != nil && id == s.created.ID {
		account := *s.created
		account.GroupIDs = append([]int64(nil), s.created.GroupIDs...)
		return &account, nil
	}
	if s.source != nil && id == s.source.ID {
		account := *s.source
		account.GroupIDs = append([]int64(nil), s.source.GroupIDs...)
		return &account, nil
	}
	return nil, ErrAccountNotFound
}

func (s *duplicateAccountRepoStub) Create(ctx context.Context, account *Account) error {
	now := time.Now().UTC()
	account.ID = 99
	account.CreatedAt = now
	account.UpdatedAt = now

	created := *account
	created.GroupIDs = append([]int64(nil), account.GroupIDs...)
	s.created = &created
	return nil
}

func (s *duplicateAccountRepoStub) BindGroups(ctx context.Context, accountID int64, groupIDs []int64) error {
	s.boundAccountID = accountID
	s.boundGroupIDs = append([]int64(nil), groupIDs...)
	if s.created != nil {
		s.created.GroupIDs = append([]int64(nil), groupIDs...)
	}
	return nil
}

func (s *duplicateAccountRepoStub) GetByIDs(ctx context.Context, ids []int64) ([]*Account, error) {
	return nil, nil
}

func (s *duplicateAccountRepoStub) ExistsByID(ctx context.Context, id int64) (bool, error) {
	return s.source != nil && s.source.ID == id, nil
}

func (s *duplicateAccountRepoStub) GetByCRSAccountID(ctx context.Context, crsAccountID string) (*Account, error) {
	return nil, nil
}

func (s *duplicateAccountRepoStub) FindByExtraField(ctx context.Context, key string, value any) ([]Account, error) {
	return nil, nil
}

func (s *duplicateAccountRepoStub) ListCRSAccountIDs(ctx context.Context) (map[string]int64, error) {
	return nil, nil
}

func (s *duplicateAccountRepoStub) Update(ctx context.Context, account *Account) error {
	return nil
}

func (s *duplicateAccountRepoStub) Delete(ctx context.Context, id int64) error {
	return nil
}

func (s *duplicateAccountRepoStub) List(ctx context.Context, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (s *duplicateAccountRepoStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, platform, accountType, status, search string, groupID int64, privacyMode string) ([]Account, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (s *duplicateAccountRepoStub) ListByGroup(ctx context.Context, groupID int64) ([]Account, error) {
	return nil, nil
}

func (s *duplicateAccountRepoStub) ListActive(ctx context.Context) ([]Account, error) {
	return nil, nil
}

func (s *duplicateAccountRepoStub) ListByPlatform(ctx context.Context, platform string) ([]Account, error) {
	return nil, nil
}

func (s *duplicateAccountRepoStub) UpdateLastUsed(ctx context.Context, id int64) error {
	return nil
}

func (s *duplicateAccountRepoStub) BatchUpdateLastUsed(ctx context.Context, updates map[int64]time.Time) error {
	return nil
}

func (s *duplicateAccountRepoStub) SetError(ctx context.Context, id int64, errorMsg string) error {
	return nil
}

func (s *duplicateAccountRepoStub) ClearError(ctx context.Context, id int64) error {
	return nil
}

func (s *duplicateAccountRepoStub) SetSchedulable(ctx context.Context, id int64, schedulable bool) error {
	return nil
}

func (s *duplicateAccountRepoStub) AutoPauseExpiredAccounts(ctx context.Context, now time.Time) (int64, error) {
	return 0, nil
}

func (s *duplicateAccountRepoStub) ListSchedulable(ctx context.Context) ([]Account, error) {
	return nil, nil
}

func (s *duplicateAccountRepoStub) ListSchedulableByGroupID(ctx context.Context, groupID int64) ([]Account, error) {
	return nil, nil
}

func (s *duplicateAccountRepoStub) ListSchedulableByPlatform(ctx context.Context, platform string) ([]Account, error) {
	return nil, nil
}

func (s *duplicateAccountRepoStub) ListSchedulableByGroupIDAndPlatform(ctx context.Context, groupID int64, platform string) ([]Account, error) {
	return nil, nil
}

func (s *duplicateAccountRepoStub) ListSchedulableByPlatforms(ctx context.Context, platforms []string) ([]Account, error) {
	return nil, nil
}

func (s *duplicateAccountRepoStub) ListSchedulableByGroupIDAndPlatforms(ctx context.Context, groupID int64, platforms []string) ([]Account, error) {
	return nil, nil
}

func (s *duplicateAccountRepoStub) ListSchedulableUngroupedByPlatform(ctx context.Context, platform string) ([]Account, error) {
	return nil, nil
}

func (s *duplicateAccountRepoStub) ListSchedulableUngroupedByPlatforms(ctx context.Context, platforms []string) ([]Account, error) {
	return nil, nil
}

func (s *duplicateAccountRepoStub) SetRateLimited(ctx context.Context, id int64, resetAt time.Time) error {
	return nil
}

func (s *duplicateAccountRepoStub) SetModelRateLimit(ctx context.Context, id int64, scope string, resetAt time.Time) error {
	return nil
}

func (s *duplicateAccountRepoStub) SetOverloaded(ctx context.Context, id int64, until time.Time) error {
	return nil
}

func (s *duplicateAccountRepoStub) SetTempUnschedulable(ctx context.Context, id int64, until time.Time, reason string) error {
	return nil
}

func (s *duplicateAccountRepoStub) ClearTempUnschedulable(ctx context.Context, id int64) error {
	return nil
}

func (s *duplicateAccountRepoStub) ClearRateLimit(ctx context.Context, id int64) error {
	return nil
}

func (s *duplicateAccountRepoStub) ClearAntigravityQuotaScopes(ctx context.Context, id int64) error {
	return nil
}

func (s *duplicateAccountRepoStub) ClearModelRateLimits(ctx context.Context, id int64) error {
	return nil
}

func (s *duplicateAccountRepoStub) UpdateSessionWindow(ctx context.Context, id int64, start, end *time.Time, status string) error {
	return nil
}

func (s *duplicateAccountRepoStub) UpdateExtra(ctx context.Context, id int64, updates map[string]any) error {
	return nil
}

func (s *duplicateAccountRepoStub) BulkUpdate(ctx context.Context, ids []int64, updates AccountBulkUpdate) (int64, error) {
	return 0, nil
}

func (s *duplicateAccountRepoStub) IncrementQuotaUsed(ctx context.Context, id int64, amount float64) error {
	return nil
}

func (s *duplicateAccountRepoStub) ResetQuotaUsed(ctx context.Context, id int64) error {
	return nil
}

var _ AccountRepository = (*duplicateAccountRepoStub)(nil)

func TestAdminServiceDuplicateAccount_CreatesConfigCloneWithoutRuntimeState(t *testing.T) {
	note := "production key"
	proxyID := int64(8)
	rateMultiplier := 1.25
	loadFactor := 7
	expiresAt := time.Unix(1893456000, 0).UTC()
	lastUsedAt := time.Now().UTC().Add(-time.Hour)
	resetAt := time.Now().UTC().Add(time.Hour)

	source := &Account{
		ID:       10,
		Name:     "source",
		Notes:    &note,
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key": "sk-test",
			"nested":  map[string]any{"tenant": "alpha"},
		},
		Extra: map[string]any{
			"quota_limit":                  100.0,
			"quota_used":                   30.0,
			"quota_daily_limit":            10.0,
			"quota_daily_used":             5.0,
			"quota_daily_start":            "2026-01-01T00:00:00Z",
			"quota_daily_reset_mode":       "fixed",
			"quota_daily_reset_hour":       2.0,
			"quota_daily_reset_at":         "2026-01-02T02:00:00Z",
			"quota_weekly_limit":           70.0,
			"quota_weekly_used":            30.0,
			"quota_weekly_start":           "2026-01-01T00:00:00Z",
			"quota_weekly_reset_at":        "2026-01-05T02:00:00Z",
			"model_rate_limits":            map[string]any{"gpt-5": "2026-01-01T00:00:00Z"},
			"crs_account_id":               "crs-1",
			"crs_kind":                     "openai",
			"crs_synced_at":                "2026-01-01T00:00:00Z",
			"import_source":                "codex_session",
			"imported_at":                  "2026-01-01T00:00:00Z",
			"openai_responses_mode":        "force_responses",
			"openai_responses_supported":   false,
			"openai_compact_mode":          "force_on",
			"openai_compact_supported":     true,
			"openai_compact_checked_at":    "2026-01-01T00:00:00Z",
			"codex_5h_used_percent":        99.0,
			"codex_usage_updated_at":       "2026-01-01T00:00:00Z",
			"passive_usage_sampled_at":     "2026-01-01T00:00:00Z",
			"session_window_utilization":   0.88,
			"custom_base_url":              "https://example.test/v1",
			"compact_model_mapping":        map[string]any{"gpt-5.4": "gpt-5.4-openai-compact"},
			"quota_notify_total_enabled":   true,
			"quota_notify_total_threshold": 80.0,
		},
		ProxyID:            &proxyID,
		Concurrency:        4,
		Priority:           9,
		RateMultiplier:     &rateMultiplier,
		LoadFactor:         &loadFactor,
		Status:             StatusError,
		ErrorMessage:       "boom",
		LastUsedAt:         &lastUsedAt,
		ExpiresAt:          &expiresAt,
		AutoPauseOnExpired: false,
		Schedulable:        false,
		RateLimitResetAt:   &resetAt,
		GroupIDs:           []int64{1, 2},
	}

	expectedNotes := *normalizeAccountNotes(&note)
	repo := &duplicateAccountRepoStub{source: source}
	svc := &adminServiceImpl{accountRepo: repo}

	got, err := svc.DuplicateAccount(context.Background(), source.ID)
	require.NoError(t, err)
	require.NotNil(t, got)

	require.Equal(t, int64(99), got.ID)
	require.Equal(t, "source (duplicate)", repo.created.Name)
	require.Equal(t, expectedNotes, *repo.created.Notes)
	require.Equal(t, PlatformOpenAI, repo.created.Platform)
	require.Equal(t, AccountTypeAPIKey, repo.created.Type)
	require.Equal(t, &proxyID, repo.created.ProxyID)
	require.Equal(t, 4, repo.created.Concurrency)
	require.Equal(t, 9, repo.created.Priority)
	require.Equal(t, 1.25, *repo.created.RateMultiplier)
	require.Equal(t, 7, *repo.created.LoadFactor)
	require.Equal(t, StatusActive, repo.created.Status)
	require.True(t, repo.created.Schedulable)
	require.Empty(t, repo.created.ErrorMessage)
	require.Nil(t, repo.created.LastUsedAt)
	require.Nil(t, repo.created.RateLimitResetAt)
	require.False(t, repo.created.AutoPauseOnExpired)
	require.Equal(t, expiresAt.Unix(), repo.created.ExpiresAt.Unix())
	require.Equal(t, int64(99), repo.boundAccountID)
	require.Equal(t, []int64{1, 2}, repo.boundGroupIDs)

	extra := repo.created.Extra
	require.Equal(t, 100.0, extra["quota_limit"])
	require.Equal(t, 10.0, extra["quota_daily_limit"])
	require.Equal(t, 70.0, extra["quota_weekly_limit"])
	require.Equal(t, "fixed", extra["quota_daily_reset_mode"])
	require.Equal(t, "force_responses", extra["openai_responses_mode"])
	require.Equal(t, "force_on", extra["openai_compact_mode"])
	require.Equal(t, "https://example.test/v1", extra["custom_base_url"])
	require.Equal(t, true, extra["quota_notify_total_enabled"])
	require.Contains(t, extra, "quota_daily_reset_at")
	require.NotContains(t, extra, "quota_used")
	require.NotContains(t, extra, "quota_daily_used")
	require.NotContains(t, extra, "quota_daily_start")
	require.NotContains(t, extra, "quota_weekly_used")
	require.NotContains(t, extra, "quota_weekly_start")
	require.NotContains(t, extra, "model_rate_limits")
	require.NotContains(t, extra, "crs_account_id")
	require.NotContains(t, extra, "crs_kind")
	require.NotContains(t, extra, "crs_synced_at")
	require.NotContains(t, extra, "import_source")
	require.NotContains(t, extra, "imported_at")
	require.NotContains(t, extra, "openai_responses_supported")
	require.NotContains(t, extra, "openai_compact_supported")
	require.NotContains(t, extra, "openai_compact_checked_at")
	require.NotContains(t, extra, "codex_5h_used_percent")
	require.NotContains(t, extra, "codex_usage_updated_at")
	require.NotContains(t, extra, "passive_usage_sampled_at")
	require.NotContains(t, extra, "session_window_utilization")

	repo.created.Credentials["nested"].(map[string]any)["tenant"] = "changed"
	require.Equal(t, "alpha", source.Credentials["nested"].(map[string]any)["tenant"])
	repo.created.Extra["compact_model_mapping"].(map[string]any)["gpt-5.4"] = "changed"
	require.Equal(t, "gpt-5.4-openai-compact", source.Extra["compact_model_mapping"].(map[string]any)["gpt-5.4"])
}

func TestBuildDuplicateAccountNameTruncatesToLimit(t *testing.T) {
	got := buildDuplicateAccountName(strings.Repeat("测", 120))
	require.LessOrEqual(t, len([]rune(got)), 100)
	require.True(t, strings.HasSuffix(got, " (duplicate)"))
}
