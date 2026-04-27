package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type accountWebhookSettingRepo struct {
	values map[string]string
}

func (r *accountWebhookSettingRepo) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (r *accountWebhookSettingRepo) GetValue(ctx context.Context, key string) (string, error) {
	if v, ok := r.values[key]; ok {
		return v, nil
	}
	return "", ErrSettingNotFound
}

func (r *accountWebhookSettingRepo) Set(ctx context.Context, key, value string) error {
	if r.values == nil {
		r.values = map[string]string{}
	}
	r.values[key] = value
	return nil
}

func (r *accountWebhookSettingRepo) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (r *accountWebhookSettingRepo) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (r *accountWebhookSettingRepo) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (r *accountWebhookSettingRepo) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestAccountSchedulingWebhookPayload(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Second)
	until := now.Add(15 * time.Minute)
	state := TempUnschedState{
		UntilUnix:       until.Unix(),
		TriggeredAtUnix: now.Unix(),
		StatusCode:      http.StatusForbidden,
		MatchedKeyword:  "suspended",
		RuleIndex:       2,
		ErrorMessage:    "account suspended by upstream",
	}
	rawState, err := json.Marshal(state)
	require.NoError(t, err)

	payloadBytes, err := buildAccountSchedulingWebhookPayload(&AccountSchedulingSuspensionEvent{
		Account: &Account{
			ID:                      42,
			Name:                    "primary",
			Platform:                PlatformOpenAI,
			Type:                    AccountTypeOAuth,
			Status:                  StatusActive,
			Schedulable:             true,
			GroupIDs:                []int64{7, 9},
			TempUnschedulableUntil:  &until,
			TempUnschedulableReason: string(rawState),
		},
		Suspension: AccountSchedulingSuspension{
			Reason: AccountSchedulingSuspensionReasonTempUnschedulable,
		},
		OccurredAt: now,
	})
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(payloadBytes, &payload))
	require.Equal(t, "sub2api", payload["source"])
	require.Equal(t, "account_scheduling_suspended", payload["type"])

	accountPayload := payload["account"].(map[string]any)
	require.Equal(t, float64(42), accountPayload["id"])
	require.Equal(t, "primary", accountPayload["name"])
	require.Equal(t, false, accountPayload["currently_schedulable"])

	suspensionPayload := payload["suspension"].(map[string]any)
	require.Equal(t, AccountSchedulingSuspensionReasonTempUnschedulable, suspensionPayload["reason"])
	require.Equal(t, "account suspended by upstream", suspensionPayload["message"])
	require.Equal(t, float64(http.StatusForbidden), suspensionPayload["status_code"])

	tempPayload := suspensionPayload["temp_unschedulable"].(map[string]any)
	require.Equal(t, "suspended", tempPayload["matched_keyword"])
	require.Equal(t, float64(2), tempPayload["rule_index"])
}

func TestAccountSchedulingWebhookSend(t *testing.T) {
	t.Parallel()

	var receivedMethod string
	var receivedUserAgent string
	var payload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedUserAgent = r.Header.Get("User-Agent")
		defer func() { _ = r.Body.Close() }()
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	cfg := defaultOpsEmailNotificationConfig()
	cfg.Webhook.Enabled = true
	cfg.Webhook.Endpoints = []OpsWebhookEndpoint{{Name: "primary", URL: server.URL + "/account"}}
	rawCfg, err := json.Marshal(cfg)
	require.NoError(t, err)

	svc := NewAccountSchedulingWebhookService(&accountWebhookSettingRepo{
		values: map[string]string{
			SettingKeyOpsEmailNotificationConfig: string(rawCfg),
		},
	}, &config.Config{})
	svc.cfg.Ops.Enabled = true
	svc.client = server.Client()

	svc.NotifyAccountSchedulingSuspended(context.Background(), &AccountSchedulingSuspensionEvent{
		Account: &Account{
			ID:               77,
			Name:             "quota-key",
			Platform:         PlatformAnthropic,
			Type:             AccountTypeAPIKey,
			Status:           StatusActive,
			Schedulable:      true,
			RateLimitResetAt: accountWebhookTimePtr(time.Now().Add(5 * time.Minute)),
		},
		Suspension: AccountSchedulingSuspension{
			Reason: AccountSchedulingSuspensionReasonRateLimited,
		},
		OccurredAt: time.Now().UTC(),
	})

	require.Equal(t, http.MethodPost, receivedMethod)
	require.Equal(t, "sub2api-account-webhook/1.0", receivedUserAgent)
	require.Equal(t, "account_scheduling_suspended", payload["type"])

	accountPayload := payload["account"].(map[string]any)
	require.Equal(t, float64(77), accountPayload["id"])
	suspensionPayload := payload["suspension"].(map[string]any)
	require.Equal(t, AccountSchedulingSuspensionReasonRateLimited, suspensionPayload["reason"])
}

func accountWebhookTimePtr(t time.Time) *time.Time {
	return &t
}
