package service

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
)

type quotaReferenceConfig struct {
	Model      string
	TokenPrice float64
}

func (s *AccountUsageService) SetQuotaReferenceServices(settingService *SettingService, billingService *BillingService) {
	if s == nil {
		return
	}
	s.settingService = settingService
	s.billingService = billingService
}

func (s *DashboardService) SetQuotaReferenceServices(settingService *SettingService, billingService *BillingService) {
	if s == nil {
		return
	}
	s.settingService = settingService
	s.billingService = billingService
}

func resolveQuotaReferenceConfig(ctx context.Context, settingService *SettingService, billingService *BillingService) (*quotaReferenceConfig, error) {
	if settingService == nil || billingService == nil {
		return nil, nil
	}

	settings, err := settingService.GetAllSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("load quota reference settings: %w", err)
	}

	model := strings.TrimSpace(settings.QuotaReferenceModel)
	if model == "" {
		return nil, nil
	}

	tokenPrice, err := billingService.GetReferenceTokenPrice(model)
	if err != nil {
		return nil, fmt.Errorf("resolve quota reference model %q: %w", model, err)
	}
	if tokenPrice <= 0 {
		return nil, fmt.Errorf("invalid quota reference token price for model %q", model)
	}

	return &quotaReferenceConfig{
		Model:      model,
		TokenPrice: tokenPrice,
	}, nil
}

func convertBalanceToReferenceTokens(balance, tokenPrice float64) int64 {
	if balance <= 0 || tokenPrice <= 0 {
		return 0
	}
	return int64(math.Round(balance / tokenPrice))
}

func clearUsageProgressReferenceTokens(progress *UsageProgress) {
	if progress == nil {
		return
	}
	progress.TotalTokens = 0
	progress.UsedTokens = 0
	progress.RemainingTokens = 0
}

func applyUsageProgressReferenceTokens(progress *UsageProgress, tokenPrice float64) {
	clearUsageProgressReferenceTokens(progress)
	if progress == nil || progress.WindowStats == nil || tokenPrice <= 0 {
		return
	}

	usedCost := progress.WindowStats.Cost
	usedTokens := convertBalanceToReferenceTokens(usedCost, tokenPrice)
	progress.UsedTokens = usedTokens
	if usedTokens == 0 {
		return
	}

	totalCost := usedCost
	if progress.Utilization > 0 {
		totalCost = usedCost * 100 / progress.Utilization
		if totalCost < usedCost {
			totalCost = usedCost
		}
	}

	totalTokens := convertBalanceToReferenceTokens(totalCost, tokenPrice)
	if totalTokens < usedTokens {
		totalTokens = usedTokens
	}

	progress.TotalTokens = totalTokens
	progress.RemainingTokens = totalTokens - usedTokens
}

func applyGroupSummaryReferenceTokens(summary *usagestats.GroupUsageSummary, tokenPrice float64) {
	if summary == nil {
		return
	}
	summary.FiveHourTokens = convertBalanceToReferenceTokens(summary.FiveHourBalance, tokenPrice)
	summary.WeeklyTokens = convertBalanceToReferenceTokens(summary.WeeklyBalance, tokenPrice)
}

func ProvideAccountUsageService(
	accountRepo AccountRepository,
	usageLogRepo UsageLogRepository,
	usageFetcher ClaudeUsageFetcher,
	geminiQuotaService *GeminiQuotaService,
	antigravityQuotaFetcher *AntigravityQuotaFetcher,
	cache *UsageCache,
	identityCache IdentityCache,
	tlsFPProfileService *TLSFingerprintProfileService,
	settingService *SettingService,
	billingService *BillingService,
) *AccountUsageService {
	svc := NewAccountUsageService(
		accountRepo,
		usageLogRepo,
		usageFetcher,
		geminiQuotaService,
		antigravityQuotaFetcher,
		cache,
		identityCache,
		tlsFPProfileService,
	)
	svc.SetQuotaReferenceServices(settingService, billingService)
	return svc
}

func ProvideDashboardService(
	usageRepo UsageLogRepository,
	aggRepo DashboardAggregationRepository,
	cache DashboardStatsCache,
	cfg *config.Config,
	settingService *SettingService,
	billingService *BillingService,
) *DashboardService {
	svc := NewDashboardService(usageRepo, aggRepo, cache, cfg)
	svc.SetQuotaReferenceServices(settingService, billingService)
	return svc
}
