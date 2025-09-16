package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"monitor-server/internal/model"
	"monitor-server/internal/repository"
)

// HostGroupHandler 主机组管理处理器
type HostGroupHandler struct {
	hostGroupRepo repository.HostGroupRepository
	hostRepo      repository.HostRepository
}

// NewHostGroupHandler 创建主机组管理处理器
func NewHostGroupHandler(db *gorm.DB) *HostGroupHandler {
	return &HostGroupHandler{
		hostGroupRepo: repository.NewHostGroupRepository(db),
		hostRepo:      repository.NewHostRepository(db),
	}
}

// CreateHostGroupRequest 创建主机组请求
type CreateHostGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Description string `json:"description"`
	Environment string `json:"environment"`
	Tags        string `json:"tags"`
	Enabled     *bool  `json:"enabled"`
}

// UpdateHostGroupRequest 更新主机组请求
type UpdateHostGroupRequest struct {
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Environment string `json:"environment"`
	Tags        string `json:"tags"`
	Enabled     *bool  `json:"enabled"`
}

// HostGroupListResponse 主机组列表响应
type HostGroupListResponse struct {
	Groups []model.HostGroup `json:"groups"`
	Total  int64             `json:"total"`
	Page   int               `json:"page"`
	Size   int               `json:"size"`
}

// HostGroupStatsResponse 主机组统计响应
type HostGroupStatsResponse struct {
	Stats []repository.GroupStats `json:"stats"`
	Total int                     `json:"total"`
}

// AddHostsToGroupRequest 添加主机到组请求
type AddHostsToGroupRequest struct {
	HostIDs []uint `json:"host_ids" binding:"required"`
}

// RemoveHostsFromGroupRequest 从组中移除主机请求
type RemoveHostsFromGroupRequest struct {
	HostIDs []uint `json:"host_ids" binding:"required"`
}

// CreateHostGroup 创建主机组
// @Summary 创建主机组
// @Description 创建新的主机组
// @Tags host-groups
// @Accept json
// @Produce json
// @Param group body CreateHostGroupRequest true "主机组信息"
// @Success 201 {object} model.HostGroup
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/host-groups [post]
func (h *HostGroupHandler) CreateHostGroup(c *gin.Context) {
	var req CreateHostGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group := &model.HostGroup{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Environment: req.Environment,
		Tags:        req.Tags,
		Enabled:     req.Enabled == nil || *req.Enabled,
	}

	if err := h.hostGroupRepo.Create(group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, group)
}

// GetHostGroups 获取主机组列表
// @Summary 获取主机组列表
// @Description 获取主机组列表，支持分页和筛选
// @Tags host-groups
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页大小" default(10)
// @Param environment query string false "环境筛选"
// @Param enabled query bool false "启用状态筛选"
// @Success 200 {object} HostGroupListResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/host-groups [get]
func (h *HostGroupHandler) GetHostGroups(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	environment := c.Query("environment")
	enabledStr := c.Query("enabled")

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	offset := (page - 1) * size

	var groups []model.HostGroup
	var total int64
	var err error

	if environment != "" {
		groups, err = h.hostGroupRepo.GetByEnvironment(environment)
		total = int64(len(groups))
		// 手动分页
		if offset < len(groups) {
			end := offset + size
			if end > len(groups) {
				end = len(groups)
			}
			groups = groups[offset:end]
		} else {
			groups = []model.HostGroup{}
		}
	} else if enabledStr != "" {
		enabled, _ := strconv.ParseBool(enabledStr)
		if enabled {
			groups, err = h.hostGroupRepo.GetEnabled()
		} else {
			// 需要在repository中添加GetDisabled方法，这里先用List
			groups, total, err = h.hostGroupRepo.List(offset, size)
		}
		if enabledStr == "true" {
			total = int64(len(groups))
		}
	} else {
		groups, total, err = h.hostGroupRepo.List(offset, size)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := HostGroupListResponse{
		Groups: groups,
		Total:  total,
		Page:   page,
		Size:   size,
	}

	c.JSON(http.StatusOK, response)
}

// GetHostGroup 获取单个主机组
// @Summary 获取单个主机组
// @Description 根据ID获取主机组详细信息
// @Tags host-groups
// @Accept json
// @Produce json
// @Param id path int true "主机组ID"
// @Param include query string false "包含关联数据" Enums(hosts)
// @Success 200 {object} model.HostGroup
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/host-groups/{id} [get]
func (h *HostGroupHandler) GetHostGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	include := c.Query("include")
	var group *model.HostGroup

	if include == "hosts" {
		group, err = h.hostGroupRepo.GetWithHosts(uint(id))
	} else {
		group, err = h.hostGroupRepo.GetByID(uint(id))
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Host group not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, group)
}

// UpdateHostGroup 更新主机组
// @Summary 更新主机组
// @Description 更新主机组信息
// @Tags host-groups
// @Accept json
// @Produce json
// @Param id path int true "主机组ID"
// @Param group body UpdateHostGroupRequest true "主机组信息"
// @Success 200 {object} model.HostGroup
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/host-groups/{id} [put]
func (h *HostGroupHandler) UpdateHostGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var req UpdateHostGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := h.hostGroupRepo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Host group not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// 更新字段
	if req.DisplayName != "" {
		group.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		group.Description = req.Description
	}
	if req.Environment != "" {
		group.Environment = req.Environment
	}
	if req.Tags != "" {
		group.Tags = req.Tags
	}
	if req.Enabled != nil {
		group.Enabled = *req.Enabled
	}

	if err := h.hostGroupRepo.Update(group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

// DeleteHostGroup 删除主机组
// @Summary 删除主机组
// @Description 删除主机组记录
// @Tags host-groups
// @Accept json
// @Produce json
// @Param id path int true "主机组ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/host-groups/{id} [delete]
func (h *HostGroupHandler) DeleteHostGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	// 检查主机组是否存在
	_, err = h.hostGroupRepo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Host group not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if err := h.hostGroupRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetHostGroupStats 获取主机组统计信息
// @Summary 获取主机组统计信息
// @Description 获取所有主机组的统计信息
// @Tags host-groups
// @Accept json
// @Produce json
// @Success 200 {object} HostGroupStatsResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/host-groups/stats [get]
func (h *HostGroupHandler) GetHostGroupStats(c *gin.Context) {
	stats, err := h.hostGroupRepo.GetGroupStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := HostGroupStatsResponse{
		Stats: stats,
		Total: len(stats),
	}

	c.JSON(http.StatusOK, response)
}

// GetGroupHosts 获取主机组中的主机
// @Summary 获取主机组中的主机
// @Description 获取指定主机组中的所有主机
// @Tags host-groups
// @Accept json
// @Produce json
// @Param id path int true "主机组ID"
// @Success 200 {object} []model.Host
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/host-groups/{id}/hosts [get]
func (h *HostGroupHandler) GetGroupHosts(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	// 检查主机组是否存在
	_, err = h.hostGroupRepo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Host group not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	hosts, err := h.hostGroupRepo.GetHostsByGroupID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, hosts)
}

// AddHostsToGroup 添加主机到组
// @Summary 添加主机到组
// @Description 将多个主机添加到指定主机组
// @Tags host-groups
// @Accept json
// @Produce json
// @Param id path int true "主机组ID"
// @Param request body AddHostsToGroupRequest true "主机ID列表"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/host-groups/{id}/hosts [post]
func (h *HostGroupHandler) AddHostsToGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var req AddHostsToGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查主机组是否存在
	_, err = h.hostGroupRepo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Host group not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if len(req.HostIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Host IDs cannot be empty"})
		return
	}

	// 验证所有主机是否存在
	for _, hostID := range req.HostIDs {
		_, err := h.hostRepo.GetByID(hostID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Host not found: " + strconv.FormatUint(uint64(hostID), 10)})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
	}

	if err := h.hostGroupRepo.AddHosts(uint(id), req.HostIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Hosts added to group successfully"})
}

// RemoveHostsFromGroup 从组中移除主机
// @Summary 从组中移除主机
// @Description 从指定主机组中移除多个主机
// @Tags host-groups
// @Accept json
// @Produce json
// @Param id path int true "主机组ID"
// @Param request body RemoveHostsFromGroupRequest true "主机ID列表"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/host-groups/{id}/hosts [delete]
func (h *HostGroupHandler) RemoveHostsFromGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var req RemoveHostsFromGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查主机组是否存在
	_, err = h.hostGroupRepo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Host group not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if len(req.HostIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Host IDs cannot be empty"})
		return
	}

	if err := h.hostGroupRepo.RemoveHosts(uint(id), req.HostIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Hosts removed from group successfully"})
}

// GetHostGroupsForHost 获取主机所属的组
// @Summary 获取主机所属的组
// @Description 获取指定主机所属的所有主机组
// @Tags hosts
// @Accept json
// @Produce json
// @Param id path int true "主机ID"
// @Success 200 {object} []model.HostGroup
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/hosts/{id}/groups [get]
func (h *HostGroupHandler) GetHostGroupsForHost(c *gin.Context) {
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

	groups, err := h.hostGroupRepo.GetHostGroups(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, groups)
}