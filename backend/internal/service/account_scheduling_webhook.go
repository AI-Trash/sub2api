package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

const (
	AccountSchedulingSuspensionReasonStatusError       = "status_error"
	AccountSchedulingSuspensionReasonStatusDisabled    = "status_disabled"
	AccountSchedulingSuspensionReasonManualPause       = "manual_pause"
	AccountSchedulingSuspensionReasonExpiredAutoPause  = "expired_auto_pause"
	AccountSchedulingSuspensionReasonRateLimited       = "rate_limited"
	AccountSchedulingSuspensionReasonOverloaded        = "overloaded"
	AccountSchedulingSuspensionReasonTempUnschedulable = "temp_unschedulable"
	AccountSchedulingSuspensionReasonQuotaExceeded     = "quota_exceeded"
	AccountSchedulingSuspensionReasonUnknown           = "unknown"
)

type AccountSchedulingSuspension struct {
	Reason     string         `json:"reason"`
	Message    string         `json:"message,omitempty"`
	Until      *time.Time     `json:"until,omitempty"`
	StatusCode *int           `json:"status_code,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type AccountSchedulingSuspensionEvent struct {
	Account    *Account
	Suspension AccountSchedulingSuspension
	OccurredAt time.Time
}

type AccountSchedulingSuspensionNotifier interface {
	NotifyAccountSchedulingSuspended(ctx context.Context, event *AccountSchedulingSuspensionEvent)
}

var defaultAccountSchedulingNotifier struct {
	mu       sync.RWMutex
	notifier AccountSchedulingSuspensionNotifier
}

func SetDefaultAccountSchedulingSuspensionNotifier(notifier AccountSchedulingSuspensionNotifier) {
	defaultAccountSchedulingNotifier.mu.Lock()
	defer defaultAccountSchedulingNotifier.mu.Unlock()
	defaultAccountSchedulingNotifier.notifier = notifier
}

func EmitAccountSchedulingSuspended(_ context.Context, event *AccountSchedulingSuspensionEvent) {
	if event == nil || event.Account == nil || event.Account.ID <= 0 {
		return
	}

	defaultAccountSchedulingNotifier.mu.RLock()
	notifier := defaultAccountSchedulingNotifier.notifier
	defaultAccountSchedulingNotifier.mu.RUnlock()
	if notifier == nil {
		return
	}

	cloned := *event
	if cloned.OccurredAt.IsZero() {
		cloned.OccurredAt = time.Now().UTC()
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.LegacyPrintf("service.account_scheduling_webhook", "[AccountSchedulingWebhook] panic recovered: %v", r)
			}
		}()
		notifier.NotifyAccountSchedulingSuspended(context.Background(), &cloned)
	}()
}

type AccountSchedulingWebhookService struct {
	settingRepo SettingRepository
	cfg         *config.Config

	limiter *slidingWindowLimiter
	client  *http.Client
}

func NewAccountSchedulingWebhookService(settingRepo SettingRepository, cfg *config.Config) *AccountSchedulingWebhookService {
	return &AccountSchedulingWebhookService{
		settingRepo: settingRepo,
		cfg:         cfg,
		limiter:     newSlidingWindowLimiter(0, time.Hour),
		client:      http.DefaultClient,
	}
}

func ProvideAccountSchedulingWebhookNotifier(settingRepo SettingRepository, cfg *config.Config) AccountSchedulingSuspensionNotifier {
	svc := NewAccountSchedulingWebhookService(settingRepo, cfg)
	SetDefaultAccountSchedulingSuspensionNotifier(svc)
	return svc
}

func (s *AccountSchedulingWebhookService) NotifyAccountSchedulingSuspended(ctx context.Context, event *AccountSchedulingSuspensionEvent) {
	if s == nil || event == nil || event.Account == nil || event.Account.ID <= 0 {
		return
	}
	if s.cfg != nil && !s.cfg.Ops.Enabled {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}

	notificationCfg, err := s.getNotificationConfig(ctx)
	if err != nil || notificationCfg == nil || !notificationCfg.Webhook.Enabled {
		return
	}
	if len(notificationCfg.Webhook.Endpoints) == 0 {
		return
	}

	s.limiter.SetLimit(notificationCfg.Webhook.RateLimitPerHour)

	payload, err := buildAccountSchedulingWebhookPayload(event)
	if err != nil {
		logger.LegacyPrintf("service.account_scheduling_webhook", "[AccountSchedulingWebhook] build payload failed: %v", err)
		return
	}

	client := s.client
	if client == nil {
		client = http.DefaultClient
	}

	timeout := time.Duration(notificationCfg.Webhook.TimeoutSeconds) * time.Second
	for _, endpoint := range notificationCfg.Webhook.Endpoints {
		targetURL := strings.TrimSpace(endpoint.URL)
		if targetURL == "" {
			continue
		}
		if !s.limiter.Allow(time.Now().UTC()) {
			continue
		}

		reqCtx := ctx
		var cancel context.CancelFunc
		if timeout > 0 {
			reqCtx, cancel = context.WithTimeout(ctx, timeout)
		}

		req, reqErr := http.NewRequestWithContext(reqCtx, http.MethodPost, targetURL, bytes.NewReader(payload))
		if reqErr != nil {
			logger.LegacyPrintf("service.account_scheduling_webhook", "[AccountSchedulingWebhook] build request failed (account=%d url=%q): %v", event.Account.ID, targetURL, reqErr)
			if cancel != nil {
				cancel()
			}
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "sub2api-account-webhook/1.0")

		resp, doErr := client.Do(req)
		if cancel != nil {
			cancel()
		}
		if doErr != nil {
			logger.LegacyPrintf("service.account_scheduling_webhook", "[AccountSchedulingWebhook] request failed (account=%d url=%q): %v", event.Account.ID, targetURL, doErr)
			continue
		}

		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			logger.LegacyPrintf("service.account_scheduling_webhook", "[AccountSchedulingWebhook] non-2xx response (account=%d url=%q status=%d)", event.Account.ID, targetURL, resp.StatusCode)
		}
	}
}

func (s *AccountSchedulingWebhookService) getNotificationConfig(ctx context.Context) (*OpsEmailNotificationConfig, error) {
	defaultCfg := defaultOpsEmailNotificationConfig()
	if s == nil || s.settingRepo == nil {
		return defaultCfg, nil
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyOpsEmailNotificationConfig)
	if err != nil {
		return defaultCfg, nil
	}
	cfg := &OpsEmailNotificationConfig{}
	if err := json.Unmarshal([]byte(raw), cfg); err != nil {
		return defaultCfg, nil
	}
	normalizeOpsEmailNotificationConfig(cfg)
	return cfg, nil
}

func buildAccountSchedulingWebhookPayload(event *AccountSchedulingSuspensionEvent) ([]byte, error) {
	if event == nil || event.Account == nil {
		return nil, fmt.Errorf("account scheduling event is required")
	}

	account := event.Account
	suspension := normalizeAccountSchedulingSuspension(account, event.Suspension)
	occurredAt := event.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	type webhookAccount struct {
		ID                      int64   `json:"id"`
		Name                    string  `json:"name"`
		Platform                string  `json:"platform"`
		Type                    string  `json:"type"`
		Status                  string  `json:"status"`
		ErrorMessage            string  `json:"error_message,omitempty"`
		Schedulable             bool    `json:"schedulable"`
		CurrentlySchedulable    bool    `json:"currently_schedulable"`
		AutoPauseOnExpired      bool    `json:"auto_pause_on_expired"`
		ExpiresAt               *string `json:"expires_at,omitempty"`
		GroupIDs                []int64 `json:"group_ids,omitempty"`
		RateLimitedAt           *string `json:"rate_limited_at,omitempty"`
		RateLimitResetAt        *string `json:"rate_limit_reset_at,omitempty"`
		OverloadUntil           *string `json:"overload_until,omitempty"`
		TempUnschedulableUntil  *string `json:"temp_unschedulable_until,omitempty"`
		TempUnschedulableReason string  `json:"temp_unschedulable_reason,omitempty"`
		SessionWindowStart      *string `json:"session_window_start,omitempty"`
		SessionWindowEnd        *string `json:"session_window_end,omitempty"`
		SessionWindowStatus     string  `json:"session_window_status,omitempty"`
	}
	type webhookSuspension struct {
		Reason            string            `json:"reason"`
		Message           string            `json:"message,omitempty"`
		Until             *string           `json:"until,omitempty"`
		StatusCode        *int              `json:"status_code,omitempty"`
		OccurredAt        string            `json:"occurred_at"`
		TempUnschedulable *TempUnschedState `json:"temp_unschedulable,omitempty"`
		Metadata          map[string]any    `json:"metadata,omitempty"`
	}

	payload := struct {
		Source     string            `json:"source"`
		Type       string            `json:"type"`
		SentAt     string            `json:"sent_at"`
		Account    webhookAccount    `json:"account"`
		Suspension webhookSuspension `json:"suspension"`
	}{
		Source: "sub2api",
		Type:   "account_scheduling_suspended",
		SentAt: time.Now().UTC().Format(time.RFC3339),
		Account: webhookAccount{
			ID:                      account.ID,
			Name:                    strings.TrimSpace(account.Name),
			Platform:                strings.TrimSpace(account.Platform),
			Type:                    strings.TrimSpace(account.Type),
			Status:                  strings.TrimSpace(account.Status),
			ErrorMessage:            strings.TrimSpace(account.ErrorMessage),
			Schedulable:             account.Schedulable,
			CurrentlySchedulable:    account.IsSchedulable(),
			AutoPauseOnExpired:      account.AutoPauseOnExpired,
			ExpiresAt:               formatWebhookTimePtr(account.ExpiresAt),
			GroupIDs:                account.GroupIDs,
			RateLimitedAt:           formatWebhookTimePtr(account.RateLimitedAt),
			RateLimitResetAt:        formatWebhookTimePtr(account.RateLimitResetAt),
			OverloadUntil:           formatWebhookTimePtr(account.OverloadUntil),
			TempUnschedulableUntil:  formatWebhookTimePtr(account.TempUnschedulableUntil),
			TempUnschedulableReason: strings.TrimSpace(account.TempUnschedulableReason),
			SessionWindowStart:      formatWebhookTimePtr(account.SessionWindowStart),
			SessionWindowEnd:        formatWebhookTimePtr(account.SessionWindowEnd),
			SessionWindowStatus:     strings.TrimSpace(account.SessionWindowStatus),
		},
		Suspension: webhookSuspension{
			Reason:            suspension.Reason,
			Message:           strings.TrimSpace(suspension.Message),
			Until:             formatWebhookTimePtr(suspension.Until),
			StatusCode:        suspension.StatusCode,
			OccurredAt:        occurredAt.UTC().Format(time.RFC3339),
			TempUnschedulable: parseTempUnschedState(account.TempUnschedulableReason),
			Metadata:          suspension.Metadata,
		},
	}

	return json.Marshal(payload)
}

func normalizeAccountSchedulingSuspension(account *Account, suspension AccountSchedulingSuspension) AccountSchedulingSuspension {
	if account == nil {
		if strings.TrimSpace(suspension.Reason) == "" {
			suspension.Reason = AccountSchedulingSuspensionReasonUnknown
		}
		return suspension
	}

	if strings.TrimSpace(suspension.Reason) == "" {
		suspension.Reason = inferAccountSchedulingSuspensionReason(account)
	}
	state := parseTempUnschedState(account.TempUnschedulableReason)
	if suspension.Until == nil {
		suspension.Until = inferAccountSchedulingSuspensionUntil(account, suspension.Reason)
	}
	if state != nil {
		if suspension.StatusCode == nil && state.StatusCode != 0 {
			code := state.StatusCode
			suspension.StatusCode = &code
		}
		if suspension.Until == nil && state.UntilUnix > 0 {
			until := time.Unix(state.UntilUnix, 0)
			suspension.Until = &until
		}
	}
	if strings.TrimSpace(suspension.Message) == "" {
		if state != nil && strings.TrimSpace(state.ErrorMessage) != "" {
			suspension.Message = state.ErrorMessage
		} else {
			suspension.Message = inferAccountSchedulingSuspensionMessage(account, suspension.Reason)
		}
	}

	if strings.TrimSpace(suspension.Reason) == "" {
		suspension.Reason = AccountSchedulingSuspensionReasonUnknown
	}
	return suspension
}

func inferAccountSchedulingSuspensionReason(account *Account) string {
	if account == nil {
		return AccountSchedulingSuspensionReasonUnknown
	}
	now := time.Now()
	switch {
	case account.Status == StatusError:
		return AccountSchedulingSuspensionReasonStatusError
	case account.Status == StatusDisabled:
		return AccountSchedulingSuspensionReasonStatusDisabled
	case !account.Schedulable:
		return AccountSchedulingSuspensionReasonManualPause
	case account.AutoPauseOnExpired && account.ExpiresAt != nil && !now.Before(*account.ExpiresAt):
		return AccountSchedulingSuspensionReasonExpiredAutoPause
	case account.TempUnschedulableUntil != nil && now.Before(*account.TempUnschedulableUntil):
		return AccountSchedulingSuspensionReasonTempUnschedulable
	case account.RateLimitResetAt != nil && now.Before(*account.RateLimitResetAt):
		return AccountSchedulingSuspensionReasonRateLimited
	case account.OverloadUntil != nil && now.Before(*account.OverloadUntil):
		return AccountSchedulingSuspensionReasonOverloaded
	case account.IsAPIKeyOrBedrock() && account.IsQuotaExceeded():
		return AccountSchedulingSuspensionReasonQuotaExceeded
	default:
		return AccountSchedulingSuspensionReasonUnknown
	}
}

func inferAccountSchedulingSuspensionUntil(account *Account, reason string) *time.Time {
	if account == nil {
		return nil
	}
	switch reason {
	case AccountSchedulingSuspensionReasonExpiredAutoPause:
		return account.ExpiresAt
	case AccountSchedulingSuspensionReasonRateLimited:
		return account.RateLimitResetAt
	case AccountSchedulingSuspensionReasonOverloaded:
		return account.OverloadUntil
	case AccountSchedulingSuspensionReasonTempUnschedulable:
		return account.TempUnschedulableUntil
	default:
		return nil
	}
}

func inferAccountSchedulingSuspensionMessage(account *Account, reason string) string {
	if account == nil {
		return ""
	}
	switch reason {
	case AccountSchedulingSuspensionReasonStatusError:
		return account.ErrorMessage
	case AccountSchedulingSuspensionReasonStatusDisabled:
		return "account status is disabled"
	case AccountSchedulingSuspensionReasonManualPause:
		return "account schedulable flag is false"
	case AccountSchedulingSuspensionReasonExpiredAutoPause:
		return "account expired and auto pause is enabled"
	case AccountSchedulingSuspensionReasonTempUnschedulable:
		return account.TempUnschedulableReason
	case AccountSchedulingSuspensionReasonRateLimited:
		return "account is rate limited"
	case AccountSchedulingSuspensionReasonOverloaded:
		return "account is overloaded"
	case AccountSchedulingSuspensionReasonQuotaExceeded:
		return "account quota exceeded"
	default:
		return ""
	}
}

func parseTempUnschedState(reason string) *TempUnschedState {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil
	}
	var state TempUnschedState
	if err := json.Unmarshal([]byte(reason), &state); err != nil {
		return nil
	}
	return &state
}

func formatWebhookTimePtr(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	v := t.UTC().Format(time.RFC3339)
	return &v
}
