package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSendHTTPEmailWithDeliveryConfig_Brevo(t *testing.T) {
	var request struct {
		APIKey  string
		Payload brevoEmailPayload
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request.APIKey = r.Header.Get("api-key")
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request.Payload))
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	svc := &EmailService{}
	err := svc.SendEmailWithDeliveryConfig(context.Background(), &EmailConfig{
		Provider: EmailProviderBrevo,
		HTTP: &HTTPEmailConfig{
			Provider: EmailProviderBrevo,
			APIKey:   "brevo-secret",
			APIURL:   server.URL,
			From:     "noreply@example.com",
			FromName: "Sub2API",
		},
	}, "user@example.com", "Hello", "<p>body</p>")

	require.NoError(t, err)
	require.Equal(t, "brevo-secret", request.APIKey)
	require.Equal(t, "noreply@example.com", request.Payload.Sender.Email)
	require.Equal(t, "Sub2API", request.Payload.Sender.Name)
	require.Len(t, request.Payload.To, 1)
	require.Equal(t, "user@example.com", request.Payload.To[0].Email)
	require.Equal(t, "Hello", request.Payload.Subject)
	require.Equal(t, "<p>body</p>", request.Payload.HTMLContent)
}

func TestSendHTTPEmailWithDeliveryConfig_ZeptoMail(t *testing.T) {
	var request struct {
		Authorization string
		Payload       zeptoMailEmailPayload
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request.Authorization = r.Header.Get("Authorization")
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request.Payload))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := &EmailService{}
	err := svc.SendEmailWithDeliveryConfig(context.Background(), &EmailConfig{
		Provider: EmailProviderZeptoMail,
		HTTP: &HTTPEmailConfig{
			Provider: EmailProviderZeptoMail,
			APIKey:   "zepto-secret",
			APIURL:   server.URL,
			From:     "noreply@example.com",
			FromName: "Sub2API",
		},
	}, "user@example.com", "Hello", "<p>body</p>")

	require.NoError(t, err)
	require.Equal(t, "Zoho-enczapikey zepto-secret", request.Authorization)
	require.Equal(t, "noreply@example.com", request.Payload.From.Address)
	require.Equal(t, "Sub2API", request.Payload.From.Name)
	require.Len(t, request.Payload.To, 1)
	require.Equal(t, "user@example.com", request.Payload.To[0].EmailAddress.Address)
	require.Equal(t, "Hello", request.Payload.Subject)
	require.Equal(t, "<p>body</p>", request.Payload.HTMLBody)
}

func TestNormalizeEmailProviderDefaultsToSMTP(t *testing.T) {
	require.Equal(t, EmailProviderSMTP, NormalizeEmailProvider(""))
	require.Equal(t, EmailProviderSMTP, NormalizeEmailProvider("unknown"))
	require.Equal(t, EmailProviderBrevo, NormalizeEmailProvider(" Brevo "))
	require.Equal(t, DefaultZeptoMailEmailAPIURL, DefaultEmailAPIURL(EmailProviderZeptoMail))
}
