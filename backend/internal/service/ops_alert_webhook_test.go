//go:build unit

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

type opsWebhookSettingRepo struct {
	values map[string]string
}

func (r *opsWebhookSettingRepo) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (r *opsWebhookSettingRepo) GetValue(ctx context.Context, key string) (string, error) {
	if v, ok := r.values[key]; ok {
		return v, nil
	}
	return "", ErrSettingNotFound
}

func (r *opsWebhookSettingRepo) Set(ctx context.Context, key, value string) error {
	if r.values == nil {
		r.values = map[string]string{}
	}
	r.values[key] = value
	return nil
}

func (r *opsWebhookSettingRepo) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (r *opsWebhookSettingRepo) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (r *opsWebhookSettingRepo) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (r *opsWebhookSettingRepo) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

type webhookTrackingOpsRepo struct {
	opsRepoMock
	updatedEventID   int64
	updatedEventSent bool
}

func (r *webhookTrackingOpsRepo) UpdateAlertEventWebhookSent(ctx context.Context, eventID int64, webhookSent bool) error {
	r.updatedEventID = eventID
	r.updatedEventSent = webhookSent
	return nil
}

func TestMaybeSendAlertWebhook(t *testing.T) {
	t.Parallel()

	var receivedMethod string
	var receivedContentType string
	var payload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedContentType = r.Header.Get("Content-Type")
		defer func() { _ = r.Body.Close() }()
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	cfg := defaultOpsEmailNotificationConfig()
	cfg.Webhook.Enabled = true
	cfg.Webhook.Endpoints = []OpsWebhookEndpoint{{Name: "primary", URL: server.URL + "/ops-alert"}}
	rawCfg, err := json.Marshal(cfg)
	require.NoError(t, err)

	settingRepo := &opsWebhookSettingRepo{
		values: map[string]string{
			SettingKeyOpsEmailNotificationConfig: string(rawCfg),
		},
	}
	opsSvc := &OpsService{
		settingRepo: settingRepo,
		cfg:         &config.Config{},
	}
	opsRepo := &webhookTrackingOpsRepo{}
	svc := &OpsAlertEvaluatorService{
		opsService:     opsSvc,
		opsRepo:        opsRepo,
		webhookLimiter: newSlidingWindowLimiter(0, time.Hour),
		webhookClient:  server.Client(),
	}

	now := time.Now().UTC().Truncate(time.Second)
	rule := &OpsAlertRule{
		ID:            11,
		Name:          "High error rate",
		Description:   "error rate is too high",
		Severity:      "P1",
		MetricType:    "error_rate",
		Operator:      ">",
		Threshold:     5,
		NotifyWebhook: true,
	}
	event := &OpsAlertEvent{
		ID:             29,
		RuleID:         11,
		Severity:       "P1",
		Status:         OpsAlertStatusFiring,
		Title:          "P1: High error rate",
		Description:    "error_rate > 5",
		MetricValue:    float64Ptr(7.5),
		ThresholdValue: float64Ptr(5),
		FiredAt:        now,
		CreatedAt:      now,
	}

	sent := svc.maybeSendAlertWebhook(context.Background(), nil, rule, event)
	require.True(t, sent)
	require.Equal(t, int64(29), opsRepo.updatedEventID)
	require.True(t, opsRepo.updatedEventSent)
	require.Equal(t, http.MethodPost, receivedMethod)
	require.Equal(t, "application/json", receivedContentType)
	require.Equal(t, "sub2api", payload["source"])
	require.Equal(t, "ops_alert", payload["type"])

	rulePayload, ok := payload["rule"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "High error rate", rulePayload["name"])
	require.Equal(t, "error_rate", rulePayload["metric_type"])

	eventPayload, ok := payload["event"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(29), eventPayload["id"])
	require.Equal(t, "firing", eventPayload["status"])
}

func TestValidateOpsEmailNotificationConfigWebhookURL(t *testing.T) {
	t.Parallel()

	cfg := defaultOpsEmailNotificationConfig()
	cfg.Webhook.Enabled = true
	cfg.Webhook.Endpoints = []OpsWebhookEndpoint{
		{Name: "broken", URL: "://bad"},
	}

	err := validateOpsEmailNotificationConfig(cfg, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "webhook.endpoints[0].url")
}
