package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// GetOpenAIImagesJSONKeepaliveSettings 获取 OpenAI 图片非流式 JSON 空白 keepalive 配置
func (h *SettingHandler) GetOpenAIImagesJSONKeepaliveSettings(c *gin.Context) {
	settings, err := h.settingService.GetOpenAIImagesJSONKeepaliveSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.OpenAIImagesJSONKeepaliveSettings{
		Enabled:                  settings.Enabled,
		KeepaliveIntervalSeconds: settings.KeepaliveIntervalSeconds,
		UserAgentKeywords:        settings.UserAgentKeywords,
		HeaderMatches:            settings.HeaderMatches,
	})
}

type UpdateOpenAIImagesJSONKeepaliveSettingsRequest struct {
	Enabled                  bool     `json:"enabled"`
	KeepaliveIntervalSeconds int      `json:"keepalive_interval_seconds"`
	UserAgentKeywords        []string `json:"user_agent_keywords"`
	HeaderMatches            []string `json:"header_matches"`
}

// UpdateOpenAIImagesJSONKeepaliveSettings 更新 OpenAI 图片非流式 JSON 空白 keepalive 配置
func (h *SettingHandler) UpdateOpenAIImagesJSONKeepaliveSettings(c *gin.Context) {
	var req UpdateOpenAIImagesJSONKeepaliveSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	settings := &service.OpenAIImagesJSONKeepaliveSettings{
		Enabled:                  req.Enabled,
		KeepaliveIntervalSeconds: req.KeepaliveIntervalSeconds,
		UserAgentKeywords:        req.UserAgentKeywords,
		HeaderMatches:            req.HeaderMatches,
	}
	if err := h.settingService.SetOpenAIImagesJSONKeepaliveSettings(c.Request.Context(), settings); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	updated, err := h.settingService.GetOpenAIImagesJSONKeepaliveSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.OpenAIImagesJSONKeepaliveSettings{
		Enabled:                  updated.Enabled,
		KeepaliveIntervalSeconds: updated.KeepaliveIntervalSeconds,
		UserAgentKeywords:        updated.UserAgentKeywords,
		HeaderMatches:            updated.HeaderMatches,
	})
}
