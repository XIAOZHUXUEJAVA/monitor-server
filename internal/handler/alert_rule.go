package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"monitor-server/internal/model"
	"monitor-server/internal/repository"
)

// AlertRuleHandler 告警规则管理处理器
type AlertRuleHandler struct {
	alertRepo repository.AlertRepository
}

// NewAlertRuleHandler 创建告警规则管理处理器
func NewAlertRuleHandler(db *gorm.DB) *AlertRuleHandler {
	return &AlertRuleHandler{
		alertRepo: repository.NewAlertRepository(db),
	}
}

// UpdateAlertRuleThresholdRequest 更新告警规则阈值请求
type UpdateAlertRuleThresholdRequest struct {
	Threshold float64 `json:"threshold" binding:"required"`
}

// CreateHostAlertRuleRequest 创建主机告警规则请求
type CreateHostAlertRuleRequest struct {
	HostID      uint    `json:"host_id" binding:"required"`
	MetricType  string  `json:"metric_type" binding:"required"`
	Severity    string  `json:"severity" binding:"required"`
	Threshold   float64 `json:"threshold" binding:"required"`
	Duration    int     `json:"duration"`
	Enabled     *bool   `json:"enabled"`
}

// GetAlertRules 获取告警规则列表
// @Summary 获取告警规则列表
// @Description 获取所有告警规则或指定主机的规则
// @Tags alert-rules
// @Accept json
// @Produce json
// @Param host_id query int false "主机ID，不提供则返回所有规则"
// @Success 200 {array} model.AlertRule
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/alert-rules [get]
func (h *AlertRuleHandler) GetAlertRules(c *gin.Context) {
	hostIDStr := c.Query("host_id")
	
	var rules []model.AlertRule
	var err error
	
	if hostIDStr != "" {
		// 获取指定主机的规则（包括全局规则）
		if hostID, parseErr := strconv.ParseUint(hostIDStr, 10, 32); parseErr == nil {
			hostIDPtr := uint(hostID)
			rules, err = h.alertRepo.GetRulesByHostID(&hostIDPtr)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host_id parameter"})
			return
		}
	} else {
		// 获取所有规则
		rules, err = h.alertRepo.GetAllRules()
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rules)
}

// UpdateAlertRuleThreshold 更新告警规则阈值
// @Summary 更新告警规则阈值
// @Description 根据指标类型和严重级别更新告警规则阈值
// @Tags alert-rules
// @Accept json
// @Produce json
// @Param metric_type path string true "指标类型" Enums(cpu,memory,disk)
// @Param severity path string true "严重级别" Enums(warning,critical)
// @Param request body UpdateAlertRuleThresholdRequest true "阈值信息"
// @Success 200 {object} model.AlertRule
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/alert-rules/{metric_type}/{severity}/threshold [put]
func (h *AlertRuleHandler) UpdateAlertRuleThreshold(c *gin.Context) {
	metricType := c.Param("metric_type")
	severity := c.Param("severity")

	var req UpdateAlertRuleThresholdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找对应的告警规则
	rules, err := h.alertRepo.GetAllRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var targetRule *model.AlertRule
	for _, rule := range rules {
		if rule.MetricType == metricType && rule.Severity == severity {
			targetRule = &rule
			break
		}
	}

	if targetRule == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert rule not found"})
		return
	}

	// 更新阈值
	targetRule.Threshold = req.Threshold
	if err := h.alertRepo.UpdateRule(targetRule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, targetRule)
}

// CreateHostAlertRule 为主机创建特定的告警规则
// @Summary 为主机创建特定的告警规则
// @Description 为主机创建或更新特定的告警规则（如果已存在则更新）
// @Tags alert-rules
// @Accept json
// @Produce json
// @Param request body CreateHostAlertRuleRequest true "主机告警规则信息"
// @Success 201 {object} model.AlertRule
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/alert-rules/host [post]
func (h *AlertRuleHandler) CreateHostAlertRule(c *gin.Context) {
	var req CreateHostAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 首先检查该主机是否已经有相同类型和严重级别的规则
	allRules, err := h.alertRepo.GetAllRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查找是否已存在相同的主机特定规则
	var existingRule *model.AlertRule
	for _, rule := range allRules {
		if rule.HostID != nil && *rule.HostID == req.HostID && 
		   rule.MetricType == req.MetricType && rule.Severity == req.Severity {
			existingRule = &rule
			break
		}
	}

	if existingRule != nil {
		// 更新现有规则
		existingRule.Threshold = req.Threshold
		if req.Duration > 0 {
			existingRule.Duration = req.Duration
		}
		if req.Enabled != nil {
			existingRule.Enabled = *req.Enabled
		}

		if err := h.alertRepo.UpdateRule(existingRule); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, existingRule)
		return
	}

	// 如果不存在，查找对应的全局规则作为模板
	var templateRule *model.AlertRule
	for _, rule := range allRules {
		if rule.HostID == nil && rule.MetricType == req.MetricType && rule.Severity == req.Severity {
			templateRule = &rule
			break
		}
	}

	if templateRule == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No global template rule found for this metric type and severity"})
		return
	}

	// 创建主机特定的规则
	newRule := &model.AlertRule{
		Name:        fmt.Sprintf("%s (主机ID: %d)", templateRule.Name, req.HostID),
		MetricType:  req.MetricType,
		Operator:    templateRule.Operator,
		Threshold:   req.Threshold,
		Duration:    req.Duration,
		Severity:    req.Severity,
		Enabled:     req.Enabled != nil && *req.Enabled,
		Description: fmt.Sprintf("主机ID %d 的自定义规则: %s", req.HostID, templateRule.Description),
		HostID:      &req.HostID,
	}

	// 如果没有提供 duration，使用模板规则的 duration
	if req.Duration == 0 {
		newRule.Duration = templateRule.Duration
	}

	// 如果没有提供 enabled，默认启用
	if req.Enabled == nil {
		newRule.Enabled = true
	}

	if err := h.alertRepo.CreateRule(newRule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newRule)
}