package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"monitor-server/internal/model"
	"monitor-server/internal/repository"
)

// HostHandler 主机管理处理器
type HostHandler struct {
	hostRepo       repository.HostRepository
	hostConfigRepo repository.HostConfigRepository
	hostGroupRepo  repository.HostGroupRepository
}

// NewHostHandler 创建主机管理处理器
func NewHostHandler(db *gorm.DB) *HostHandler {
	return &HostHandler{
		hostRepo:       repository.NewHostRepository(db),
		hostConfigRepo: repository.NewHostConfigRepository(db),
		hostGroupRepo:  repository.NewHostGroupRepository(db),
	}
}

// CreateHostRequest 创建主机请求
type CreateHostRequest struct {
	Hostname          string `json:"hostname" binding:"required"`
	DisplayName       string `json:"display_name" binding:"required"`
	IPAddress         string `json:"ip_address" binding:"required"`
	Environment       string `json:"environment" binding:"required"`
	Location          string `json:"location"`
	Tags              string `json:"tags"`
	Description       string `json:"description"`
	MonitoringEnabled *bool  `json:"monitoring_enabled"`
	OS                string `json:"os"`
	Platform          string `json:"platform"`
	CPUCores          int    `json:"cpu_cores"`
	TotalMemory       uint64 `json:"total_memory"`
	Agent             *bool  `json:"agent"`
}

// UpdateHostRequest 更新主机请求
type UpdateHostRequest struct {
	DisplayName       string `json:"display_name"`
	IPAddress         string `json:"ip_address"`
	Environment       string `json:"environment"`
	Location          string `json:"location"`
	Tags              string `json:"tags"`
	Description       string `json:"description"`
	Status            string `json:"status"`
	MonitoringEnabled *bool  `json:"monitoring_enabled"`
	OS                string `json:"os"`
	Platform          string `json:"platform"`
	CPUCores          int    `json:"cpu_cores"`
	TotalMemory       uint64 `json:"total_memory"`
	Agent             *bool  `json:"agent"`
}

// HostListResponse 主机列表响应
type HostListResponse struct {
	Hosts []model.Host `json:"hosts"`
	Total int64        `json:"total"`
	Page  int          `json:"page"`
	Size  int          `json:"size"`
}

// HostStatsResponse 主机统计响应
type HostStatsResponse struct {
	StatusStats      map[string]int64 `json:"status_stats"`
	EnvironmentStats map[string]int64 `json:"environment_stats"`
	TotalHosts       int64            `json:"total_hosts"`
	OnlineHosts      int64            `json:"online_hosts"`
}

// CreateHost 创建主机
// @Summary 创建主机
// @Description 创建新的主机记录
// @Tags hosts
// @Accept json
// @Produce json
// @Param host body CreateHostRequest true "主机信息"
// @Success 201 {object} model.Host
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/hosts [post]
func (h *HostHandler) CreateHost(c *gin.Context) {
	var req CreateHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	host := &model.Host{
		Hostname:          req.Hostname,
		DisplayName:       req.DisplayName,
		IPAddress:         req.IPAddress,
		Environment:       req.Environment,
		Location:          req.Location,
		Tags:              req.Tags,
		Description:       req.Description,
		Status:            "unknown",
		MonitoringEnabled: req.MonitoringEnabled != nil && *req.MonitoringEnabled,
		OS:                req.OS,
		Platform:          req.Platform,
		CPUCores:          req.CPUCores,
		TotalMemory:       req.TotalMemory,
		Agent:             req.Agent != nil && *req.Agent,
	}

	if err := h.hostRepo.Create(host); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, host)
}

// GetHosts 获取主机列表
// @Summary 获取主机列表
// @Description 获取主机列表，支持分页和筛选
// @Tags hosts
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页大小" default(10)
// @Param environment query string false "环境筛选"
// @Param status query string false "状态筛选"
// @Param keyword query string false "关键词搜索"
// @Success 200 {object} HostListResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/hosts [get]
func (h *HostHandler) GetHosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	environment := c.Query("environment")
	status := c.Query("status")
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	offset := (page - 1) * size

	var hosts []model.Host
	var total int64
	var err error

	if keyword != "" || environment != "" || status != "" {
		hosts, total, err = h.hostRepo.Search(keyword, environment, status, offset, size)
	} else {
		hosts, total, err = h.hostRepo.List(offset, size)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := HostListResponse{
		Hosts: hosts,
		Total: total,
		Page:  page,
		Size:  size,
	}

	c.JSON(http.StatusOK, response)
}

// GetHost 获取单个主机
// @Summary 获取单个主机
// @Description 根据ID获取主机详细信息
// @Tags hosts
// @Accept json
// @Produce json
// @Param id path int true "主机ID"
// @Param include query string false "包含关联数据" Enums(configs, groups, all)
// @Success 200 {object} model.Host
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/hosts/{id} [get]
func (h *HostHandler) GetHost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host ID"})
		return
	}

	include := c.Query("include")
	var host *model.Host

	switch include {
	case "configs":
		host, err = h.hostRepo.GetWithConfigs(uint(id))
	case "groups":
		host, err = h.hostRepo.GetWithGroups(uint(id))
	case "all":
		host, err = h.hostRepo.GetWithAll(uint(id))
	default:
		host, err = h.hostRepo.GetByID(uint(id))
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Host not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, host)
}

// UpdateHost 更新主机
// @Summary 更新主机
// @Description 更新主机信息
// @Tags hosts
// @Accept json
// @Produce json
// @Param id path int true "主机ID"
// @Param host body UpdateHostRequest true "主机信息"
// @Success 200 {object} model.Host
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/hosts/{id} [put]
func (h *HostHandler) UpdateHost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host ID"})
		return
	}

	var req UpdateHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	host, err := h.hostRepo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Host not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// 更新字段
	if req.DisplayName != "" {
		host.DisplayName = req.DisplayName
	}
	if req.IPAddress != "" {
		host.IPAddress = req.IPAddress
	}
	if req.Environment != "" {
		host.Environment = req.Environment
	}
	if req.Location != "" {
		host.Location = req.Location
	}
	if req.Tags != "" {
		host.Tags = req.Tags
	}
	if req.Description != "" {
		host.Description = req.Description
	}
	if req.Status != "" {
		host.Status = req.Status
	}
	if req.MonitoringEnabled != nil {
		host.MonitoringEnabled = *req.MonitoringEnabled
	}
	if req.OS != "" {
		host.OS = req.OS
	}
	if req.Platform != "" {
		host.Platform = req.Platform
	}
	if req.CPUCores > 0 {
		host.CPUCores = req.CPUCores
	}
	if req.TotalMemory > 0 {
		host.TotalMemory = req.TotalMemory
	}
	if req.Agent != nil {
		host.Agent = *req.Agent
	}

	if err := h.hostRepo.Update(host); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, host)
}

// DeleteHost 删除主机
// @Summary 删除主机
// @Description 删除主机记录
// @Tags hosts
// @Accept json
// @Produce json
// @Param id path int true "主机ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/hosts/{id} [delete]
func (h *HostHandler) DeleteHost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host ID"})
		return
	}

	// 检查主机是否存在
	_, err = h.hostRepo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Host not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if err := h.hostRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetHostStats 获取主机统计信息
// @Summary 获取主机统计信息
// @Description 获取主机状态和环境统计信息
// @Tags hosts
// @Accept json
// @Produce json
// @Success 200 {object} HostStatsResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/hosts/stats [get]
func (h *HostHandler) GetHostStats(c *gin.Context) {
	statusStats, err := h.hostRepo.CountByStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	environmentStats, err := h.hostRepo.CountByEnvironment()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var totalHosts int64
	var onlineHosts int64
	for _, count := range statusStats {
		totalHosts += count
	}
	if count, ok := statusStats["online"]; ok {
		onlineHosts = count
	}

	response := HostStatsResponse{
		StatusStats:      statusStats,
		EnvironmentStats: environmentStats,
		TotalHosts:       totalHosts,
		OnlineHosts:      onlineHosts,
	}

	c.JSON(http.StatusOK, response)
}

// BatchUpdateHostStatus 批量更新主机状态
// @Summary 批量更新主机状态
// @Description 批量更新多个主机的状态
// @Tags hosts
// @Accept json
// @Produce json
// @Param request body BatchUpdateStatusRequest true "批量更新请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/hosts/batch/status [put]
func (h *HostHandler) BatchUpdateHostStatus(c *gin.Context) {
	var req BatchUpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.HostIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Host IDs cannot be empty"})
		return
	}

	if err := h.hostRepo.BatchUpdateStatus(req.HostIDs, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Host status updated successfully"})
}

// BatchUpdateStatusRequest 批量更新状态请求
type BatchUpdateStatusRequest struct {
	HostIDs []uint `json:"host_ids" binding:"required"`
	Status  string `json:"status" binding:"required"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Message string `json:"message"`
}