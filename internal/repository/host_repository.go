package repository

import (
	"time"

	"gorm.io/gorm"

	"monitor-server/internal/model"
)

// HostRepository 主机仓库接口
type HostRepository interface {
	// 基本CRUD操作
	Create(host *model.Host) error
	GetByID(id uint) (*model.Host, error)
	GetByHostname(hostname string) (*model.Host, error)
	Update(host *model.Host) error
	Delete(id uint) error
	List(offset, limit int) ([]model.Host, int64, error)
	
	// 条件查询
	GetByEnvironment(environment string) ([]model.Host, error)
	GetByStatus(status string) ([]model.Host, error)
	GetOnlineHosts() ([]model.Host, error)
	GetMonitoringEnabledHosts() ([]model.Host, error)
	
	// 关联查询
	GetWithConfigs(id uint) (*model.Host, error)
	GetWithGroups(id uint) (*model.Host, error)
	GetWithAll(id uint) (*model.Host, error)
	
	// 批量操作
	BatchUpdateStatus(hostIDs []uint, status string) error
	BatchToggleMonitoring(hostIDs []uint, enabled bool) error
	
	// 统计查询
	CountByStatus() (map[string]int64, error)
	CountByEnvironment() (map[string]int64, error)
	
	// 高级查询
	Search(keyword string, environment string, status string, offset, limit int) ([]model.Host, int64, error)
	UpdateLastSeen(hostname string) error
}

// HostConfigRepository 主机配置仓库接口
type HostConfigRepository interface {
	Create(config *model.HostConfig) error
	GetByID(id uint) (*model.HostConfig, error)
	GetByHostID(hostID uint) ([]model.HostConfig, error)
	GetByHostIDAndKey(hostID uint, key string) (*model.HostConfig, error)
	Update(config *model.HostConfig) error
	Delete(id uint) error
	DeleteByHostID(hostID uint) error
	
	// 按分类查询
	GetByCategory(hostID uint, category string) ([]model.HostConfig, error)
	
	// 批量操作
	CreateBatch(configs []model.HostConfig) error
	UpdateValue(hostID uint, key, value string) error
}

// HostGroupRepository 主机组仓库接口
type HostGroupRepository interface {
	Create(group *model.HostGroup) error
	GetByID(id uint) (*model.HostGroup, error)
	GetByName(name string) (*model.HostGroup, error)
	Update(group *model.HostGroup) error
	Delete(id uint) error
	List(offset, limit int) ([]model.HostGroup, int64, error)
	
	// 条件查询
	GetByEnvironment(environment string) ([]model.HostGroup, error)
	GetEnabled() ([]model.HostGroup, error)
	
	// 关联查询
	GetWithHosts(id uint) (*model.HostGroup, error)
	GetHostsByGroupID(groupID uint) ([]model.Host, error)
	
	// 主机组成员管理
	AddHost(groupID, hostID uint) error
	RemoveHost(groupID, hostID uint) error
	IsHostInGroup(groupID, hostID uint) (bool, error)
	GetHostGroups(hostID uint) ([]model.HostGroup, error)
	
	// 批量操作
	AddHosts(groupID uint, hostIDs []uint) error
	RemoveHosts(groupID uint, hostIDs []uint) error
	
	// 统计查询
	CountHosts(groupID uint) (int64, error)
	GetGroupStats() ([]GroupStats, error)
}

// GroupStats 主机组统计信息
type GroupStats struct {
	GroupID     uint   `json:"group_id"`
	GroupName   string `json:"group_name"`
	HostCount   int64  `json:"host_count"`
	OnlineCount int64  `json:"online_count"`
	Environment string `json:"environment"`
}

// hostRepository GORM实现
type hostRepository struct {
	db *gorm.DB
}

// NewHostRepository 创建主机仓库
func NewHostRepository(db *gorm.DB) HostRepository {
	return &hostRepository{db: db}
}

func (r *hostRepository) Create(host *model.Host) error {
	return r.db.Create(host).Error
}

func (r *hostRepository) GetByID(id uint) (*model.Host, error) {
	var host model.Host
	err := r.db.First(&host, id).Error
	if err != nil {
		return nil, err
	}
	return &host, nil
}

func (r *hostRepository) GetByHostname(hostname string) (*model.Host, error) {
	var host model.Host
	err := r.db.Where("hostname = ?", hostname).First(&host).Error
	if err != nil {
		return nil, err
	}
	return &host, nil
}

func (r *hostRepository) Update(host *model.Host) error {
	return r.db.Save(host).Error
}

func (r *hostRepository) Delete(id uint) error {
	return r.db.Delete(&model.Host{}, id).Error
}

func (r *hostRepository) List(offset, limit int) ([]model.Host, int64, error) {
	var hosts []model.Host
	var total int64
	
	err := r.db.Model(&model.Host{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	err = r.db.Offset(offset).Limit(limit).Find(&hosts).Error
	return hosts, total, err
}

func (r *hostRepository) GetByEnvironment(environment string) ([]model.Host, error) {
	var hosts []model.Host
	err := r.db.Where("environment = ?", environment).Find(&hosts).Error
	return hosts, err
}

func (r *hostRepository) GetByStatus(status string) ([]model.Host, error) {
	var hosts []model.Host
	err := r.db.Where("status = ?", status).Find(&hosts).Error
	return hosts, err
}

func (r *hostRepository) GetOnlineHosts() ([]model.Host, error) {
	return r.GetByStatus("online")
}

func (r *hostRepository) GetMonitoringEnabledHosts() ([]model.Host, error) {
	var hosts []model.Host
	err := r.db.Where("monitoring_enabled = ?", true).Find(&hosts).Error
	return hosts, err
}

func (r *hostRepository) GetWithConfigs(id uint) (*model.Host, error) {
	var host model.Host
	err := r.db.Preload("Configs").First(&host, id).Error
	if err != nil {
		return nil, err
	}
	return &host, nil
}

func (r *hostRepository) GetWithGroups(id uint) (*model.Host, error) {
	var host model.Host
	err := r.db.Preload("Groups").First(&host, id).Error
	if err != nil {
		return nil, err
	}
	return &host, nil
}

func (r *hostRepository) GetWithAll(id uint) (*model.Host, error) {
	var host model.Host
	err := r.db.Preload("Configs").Preload("Groups").First(&host, id).Error
	if err != nil {
		return nil, err
	}
	return &host, nil
}

func (r *hostRepository) BatchUpdateStatus(hostIDs []uint, status string) error {
	return r.db.Model(&model.Host{}).Where("id IN ?", hostIDs).Update("status", status).Error
}

func (r *hostRepository) BatchToggleMonitoring(hostIDs []uint, enabled bool) error {
	return r.db.Model(&model.Host{}).Where("id IN ?", hostIDs).Update("monitoring_enabled", enabled).Error
}

func (r *hostRepository) CountByStatus() (map[string]int64, error) {
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	
	var results []StatusCount
	err := r.db.Model(&model.Host{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	statusMap := make(map[string]int64)
	for _, result := range results {
		statusMap[result.Status] = result.Count
	}
	
	return statusMap, nil
}

func (r *hostRepository) CountByEnvironment() (map[string]int64, error) {
	type EnvCount struct {
		Environment string `json:"environment"`
		Count       int64  `json:"count"`
	}
	
	var results []EnvCount
	err := r.db.Model(&model.Host{}).
		Select("environment, COUNT(*) as count").
		Group("environment").
		Scan(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	envMap := make(map[string]int64)
	for _, result := range results {
		envMap[result.Environment] = result.Count
	}
	
	return envMap, nil
}

func (r *hostRepository) Search(keyword string, environment string, status string, offset, limit int) ([]model.Host, int64, error) {
	query := r.db.Model(&model.Host{})
	
	if keyword != "" {
		query = query.Where("hostname ILIKE ? OR display_name ILIKE ? OR ip_address ILIKE ?", 
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	
	if environment != "" {
		query = query.Where("environment = ?", environment)
	}
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	var hosts []model.Host
	err = query.Offset(offset).Limit(limit).Find(&hosts).Error
	return hosts, total, err
}

func (r *hostRepository) UpdateLastSeen(hostname string) error {
	return r.db.Model(&model.Host{}).
		Where("hostname = ?", hostname).
		Update("last_seen", time.Now()).Error
}

// hostConfigRepository GORM实现
type hostConfigRepository struct {
	db *gorm.DB
}

// NewHostConfigRepository 创建主机配置仓库
func NewHostConfigRepository(db *gorm.DB) HostConfigRepository {
	return &hostConfigRepository{db: db}
}

func (r *hostConfigRepository) Create(config *model.HostConfig) error {
	return r.db.Create(config).Error
}

func (r *hostConfigRepository) GetByID(id uint) (*model.HostConfig, error) {
	var config model.HostConfig
	err := r.db.First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *hostConfigRepository) GetByHostID(hostID uint) ([]model.HostConfig, error) {
	var configs []model.HostConfig
	err := r.db.Where("host_id = ?", hostID).Find(&configs).Error
	return configs, err
}

func (r *hostConfigRepository) GetByHostIDAndKey(hostID uint, key string) (*model.HostConfig, error) {
	var config model.HostConfig
	err := r.db.Where("host_id = ? AND key = ?", hostID, key).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *hostConfigRepository) Update(config *model.HostConfig) error {
	return r.db.Save(config).Error
}

func (r *hostConfigRepository) Delete(id uint) error {
	return r.db.Delete(&model.HostConfig{}, id).Error
}

func (r *hostConfigRepository) DeleteByHostID(hostID uint) error {
	return r.db.Where("host_id = ?", hostID).Delete(&model.HostConfig{}).Error
}

func (r *hostConfigRepository) GetByCategory(hostID uint, category string) ([]model.HostConfig, error) {
	var configs []model.HostConfig
	err := r.db.Where("host_id = ? AND category = ?", hostID, category).Find(&configs).Error
	return configs, err
}

func (r *hostConfigRepository) CreateBatch(configs []model.HostConfig) error {
	return r.db.CreateInBatches(configs, 100).Error
}

func (r *hostConfigRepository) UpdateValue(hostID uint, key, value string) error {
	return r.db.Model(&model.HostConfig{}).
		Where("host_id = ? AND key = ?", hostID, key).
		Update("value", value).Error
}

// hostGroupRepository GORM实现
type hostGroupRepository struct {
	db *gorm.DB
}

// NewHostGroupRepository 创建主机组仓库
func NewHostGroupRepository(db *gorm.DB) HostGroupRepository {
	return &hostGroupRepository{db: db}
}

func (r *hostGroupRepository) Create(group *model.HostGroup) error {
	return r.db.Create(group).Error
}

func (r *hostGroupRepository) GetByID(id uint) (*model.HostGroup, error) {
	var group model.HostGroup
	err := r.db.First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *hostGroupRepository) GetByName(name string) (*model.HostGroup, error) {
	var group model.HostGroup
	err := r.db.Where("name = ?", name).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *hostGroupRepository) Update(group *model.HostGroup) error {
	return r.db.Save(group).Error
}

func (r *hostGroupRepository) Delete(id uint) error {
	return r.db.Delete(&model.HostGroup{}, id).Error
}

func (r *hostGroupRepository) List(offset, limit int) ([]model.HostGroup, int64, error) {
	var groups []model.HostGroup
	var total int64
	
	err := r.db.Model(&model.HostGroup{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	err = r.db.Offset(offset).Limit(limit).Find(&groups).Error
	return groups, total, err
}

func (r *hostGroupRepository) GetByEnvironment(environment string) ([]model.HostGroup, error) {
	var groups []model.HostGroup
	err := r.db.Where("environment = ?", environment).Find(&groups).Error
	return groups, err
}

func (r *hostGroupRepository) GetEnabled() ([]model.HostGroup, error) {
	var groups []model.HostGroup
	err := r.db.Where("enabled = ?", true).Find(&groups).Error
	return groups, err
}

func (r *hostGroupRepository) GetWithHosts(id uint) (*model.HostGroup, error) {
	var group model.HostGroup
	err := r.db.Preload("Hosts").First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *hostGroupRepository) GetHostsByGroupID(groupID uint) ([]model.Host, error) {
	var hosts []model.Host
	err := r.db.Joins("JOIN host_group_members ON hosts.id = host_group_members.host_id").
		Where("host_group_members.host_group_id = ?", groupID).
		Find(&hosts).Error
	return hosts, err
}

func (r *hostGroupRepository) AddHost(groupID, hostID uint) error {
	member := model.HostGroupMember{
		HostID:      hostID,
		HostGroupID: groupID,
		JoinedAt:    time.Now(),
		Role:        "member",
	}
	return r.db.Create(&member).Error
}

func (r *hostGroupRepository) RemoveHost(groupID, hostID uint) error {
	return r.db.Where("host_group_id = ? AND host_id = ?", groupID, hostID).
		Delete(&model.HostGroupMember{}).Error
}

func (r *hostGroupRepository) IsHostInGroup(groupID, hostID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.HostGroupMember{}).
		Where("host_group_id = ? AND host_id = ?", groupID, hostID).
		Count(&count).Error
	return count > 0, err
}

func (r *hostGroupRepository) GetHostGroups(hostID uint) ([]model.HostGroup, error) {
	var groups []model.HostGroup
	err := r.db.Joins("JOIN host_group_members ON host_groups.id = host_group_members.host_group_id").
		Where("host_group_members.host_id = ?", hostID).
		Find(&groups).Error
	return groups, err
}

func (r *hostGroupRepository) AddHosts(groupID uint, hostIDs []uint) error {
	var members []model.HostGroupMember
	for _, hostID := range hostIDs {
		members = append(members, model.HostGroupMember{
			HostID:      hostID,
			HostGroupID: groupID,
			JoinedAt:    time.Now(),
			Role:        "member",
		})
	}
	return r.db.CreateInBatches(members, 100).Error
}

func (r *hostGroupRepository) RemoveHosts(groupID uint, hostIDs []uint) error {
	return r.db.Where("host_group_id = ? AND host_id IN ?", groupID, hostIDs).
		Delete(&model.HostGroupMember{}).Error
}

func (r *hostGroupRepository) CountHosts(groupID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.HostGroupMember{}).
		Where("host_group_id = ?", groupID).
		Count(&count).Error
	return count, err
}

func (r *hostGroupRepository) GetGroupStats() ([]GroupStats, error) {
	var stats []GroupStats
	err := r.db.Raw(`
		SELECT 
			hg.id as group_id,
			hg.name as group_name,
			hg.environment,
			COUNT(hgm.host_id) as host_count,
			COUNT(CASE WHEN h.status = 'online' THEN 1 END) as online_count
		FROM host_groups hg
		LEFT JOIN host_group_members hgm ON hg.id = hgm.host_group_id
		LEFT JOIN hosts h ON hgm.host_id = h.id
		WHERE hg.enabled = true
		GROUP BY hg.id, hg.name, hg.environment
		ORDER BY hg.name
	`).Scan(&stats).Error
	
	return stats, err
}