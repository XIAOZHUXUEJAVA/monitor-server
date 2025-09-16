package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"monitor-server/internal/model"
	"monitor-server/internal/repository"
)

// HostConfigHandler 主机配置管理处理器
type HostConfigHandler struct {
	hostConfigRepo repository.HostConfigRepository
	hostRepo       repository.HostRepository
}

// NewHostConfigHandler 创建主机配置管理处理器
func NewHostConfigHandler(db *gorm.DB) *HostConfigHandler {
	return &HostConfigHandler{
		hostConfigRepo: repository.NewHostConfigRepository(db),
		hostRepo:       repository.NewHostRepository(db),
	}
}

// CreateHostConfigRequest 创建主机配置请求
type CreateHostConfigRequest struct {
	HostID      uint   `json:"host_id" binding:"required"`
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Category    string `json:"category" binding:"required"`
	Description string `json:"description"`
	Editable    *bool  `json:"editable"`
}

// UpdateHostConfigRequest 更新主机配置请求
type UpdateHostConfigRequest struct {
	Value       string `json:"value"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Editable    *bool  `json:"editable"`
}

// BatchCreateHostConfigRequest 批量创建主机配置请求
type BatchCreateHostConfigRequest struct {
	HostID  uint                      `json:"host_id" binding:"required"`
	Configs []CreateConfigItemRequest `json:"configs" binding:"required"`
}

// CreateConfigItemRequest 创建配置项请求
type CreateConfigItemRequest struct {
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Category    string `json:"category" binding:"required"`
	Description string `json:"description"`
	Editable    *bool  `json:"editable"`
}

// HostConfigListResponse 主机配置列表响应
type HostConfigListResponse struct {
	Configs []model.HostConfig `json:"configs"`
	Total   int                `json:"total"`
}

// CreateHostConfig 创建主机配置
// @Summary 创建主机配置
// @Description 为指定主机创建新的配置项
// @Tags host-configs
// @Accept json
// @Produce json
// @Param config body CreateHostConfigRequest true "配置信息"
// @Success 201 {object} model.HostConfig
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/host-configs [post]
func (h *HostConfigHandler) CreateHostConfig(c *gin.Context) {
	var req CreateHostConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证主机是否存在
	_, err := h.hostRepo.GetByID(req.HostID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Host not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	config := &model.HostConfig{
		HostID:      req.HostID,
		Key:         req.Key,
		Value:       req.Value,
		Type:        req.Type,
		Category:    req.Category,
		Description: req.Description,
		Editable:    req.Editable != nil && *req.Editable,
	}

	if err := h.hostConfigRepo.Create(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, config)
}

// GetHostConfigs 获取主机配置列表
// @Summary 获取主机配置列表
// @Description 获取指定主机的配置列表
// @Tags host-configs
// @Accept json
// @Produce json
// @Param host_id path int true "主机ID"
// @Param category query string false "配置分类筛选"
// @Success 200 {object} HostConfigListResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/hosts/{host_id}/configs [get]
func (h *HostConfigHandler) GetHostConfigs(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host ID"})
		return
	}

	// 验证主机是否存在
	_, err = h.hostRepo.GetByID(uint(hostID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Host not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	category := c.Query("category")
	var configs []model.HostConfig

	if category != "" {
		configs, err = h.hostConfigRepo.GetByCategory(uint(hostID), category)
	} else {
		configs, err = h.hostConfigRepo.GetByHostID(uint(hostID))
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := HostConfigListResponse{
		Configs: configs,
		Total:   len(configs),
	}

	c.JSON(http.StatusOK, response)
}

// GetHostConfig 获取单个主机配置
// @Summary 获取单个主机配置
// @Description 根据ID获取主机配置详细信息
// @Tags host-configs
// @Accept json
// @Produce json
// @Param id path int true "配置ID"
// @Success 200 {object} model.HostConfig
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/host-configs/{id} [get]
func (h *HostConfigHandler) GetHostConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	config, err := h.hostConfigRepo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateHostConfig 更新主机配置
// @Summary 更新主机配置
// @Description 更新主机配置信息
// @Tags host-configs
// @Accept json
// @Produce json
// @Param id path int true "配置ID"
// @Param config body UpdateHostConfigRequest true "配置信息"
// @Success 200 {object} model.HostConfig
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/host-configs/{id} [put]
func (h *HostConfigHandler) UpdateHostConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	var req UpdateHostConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := h.hostConfigRepo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// 更新字段
	if req.Value != "" {
		config.Value = req.Value
	}
	if req.Type != "" {
		config.Type = req.Type
	}
	if req.Category != "" {
		config.Category = req.Category
	}
	if req.Description != "" {
		config.Description = req.Description
	}
	if req.Editable != nil {
		config.Editable = *req.Editable
	}

	if err := h.hostConfigRepo.Update(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// DeleteHostConfig 删除主机配置
// @Summary 删除主机配置
// @Description 删除主机配置记录
// @Tags host-configs
// @Accept json
// @Produce json
// @Param id path int true "配置ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/host-configs/{id} [delete]
func (h *HostConfigHandler) DeleteHostConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	// 检查配置是否存在
	_, err = h.hostConfigRepo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if err := h.hostConfigRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// BatchCreateHostConfigs 批量创建主机配置
// @Summary 批量创建主机配置
// @Description 为指定主机批量创建多个配置项
// @Tags host-configs
// @Accept json
// @Produce json
// @Param configs body BatchCreateHostConfigRequest true "批量配置信息"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/host-configs/batch [post]
func (h *HostConfigHandler) BatchCreateHostConfigs(c *gin.Context) {
	var req BatchCreateHostConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证主机是否存在
	_, err := h.hostRepo.GetByID(req.HostID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Host not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if len(req.Configs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Configs cannot be empty"})
		return
	}

	var configs []model.HostConfig
	for _, configReq := range req.Configs {
		config := model.HostConfig{
			HostID:      req.HostID,
			Key:         configReq.Key,
			Value:       configReq.Value,
			Type:        configReq.Type,
			Category:    configReq.Category,
			Description: configReq.Description,
			Editable:    configReq.Editable != nil && *configReq.Editable,
		}
		configs = append(configs, config)
	}

	if err := h.hostConfigRepo.CreateBatch(configs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Host configs created successfully"})
}

// GetHostConfigByKey 根据主机ID和配置键获取配置
// @Summary 根据主机ID和配置键获取配置
// @Description 根据主机ID和配置键获取特定配置项
// @Tags host-configs
// @Accept json
// @Produce json
// @Param host_id path int true "主机ID"
// @Param key path string true "配置键"
// @Success 200 {object} model.HostConfig
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/hosts/{host_id}/configs/{key} [get]
func (h *HostConfigHandler) GetHostConfigByKey(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host ID"})
		return
	}

	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Config key cannot be empty"})
		return
	}

	config, err := h.hostConfigRepo.GetByHostIDAndKey(uint(hostID), key)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateHostConfigValue 更新主机配置值
// @Summary 更新主机配置值
// @Description 根据主机ID和配置键更新配置值
// @Tags host-configs
// @Accept json
// @Produce json
// @Param host_id path int true "主机ID"
// @Param key path string true "配置键"
// @Param request body UpdateConfigValueRequest true "配置值"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/hosts/{host_id}/configs/{key} [put]
func (h *HostConfigHandler) UpdateHostConfigValue(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host ID"})
		return
	}

	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Config key cannot be empty"})
		return
	}

	var req UpdateConfigValueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查配置是否存在
	_, err = h.hostConfigRepo.GetByHostIDAndKey(uint(hostID), key)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if err := h.hostConfigRepo.UpdateValue(uint(hostID), key, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config value updated successfully"})
}

// UpdateConfigValueRequest 更新配置值请求
type UpdateConfigValueRequest struct {
	Value string `json:"value" binding:"required"`
}