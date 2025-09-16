package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型，包含通用字段
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// SystemMetrics 系统指标模型
type SystemMetrics struct {
	BaseModel
	Hostname     string    `gorm:"type:varchar(255);not null;index" json:"hostname"`
	CPUUsage     float64   `gorm:"type:decimal(5,2);not null" json:"cpu_usage"`
	MemoryUsage  float64   `gorm:"type:decimal(5,2);not null" json:"memory_usage"`
	MemoryTotal  uint64    `gorm:"not null" json:"memory_total"`
	MemoryUsed   uint64    `gorm:"not null" json:"memory_used"`
	DiskUsage    float64   `gorm:"type:decimal(5,2);not null" json:"disk_usage"`
	DiskTotal    uint64    `gorm:"not null" json:"disk_total"`
	DiskUsed     uint64    `gorm:"not null" json:"disk_used"`
	NetworkSent  uint64    `gorm:"not null" json:"network_sent"`
	NetworkRecv  uint64    `gorm:"not null" json:"network_recv"`
	Timestamp    time.Time `gorm:"not null;index" json:"timestamp"`
}

// SystemInfo 系统信息模型（用于数据库存储）
type SystemInfoDB struct {
	BaseModel
	Hostname        string `gorm:"type:varchar(255);not null;uniqueIndex" json:"hostname"`
	OS              string `gorm:"type:varchar(100);not null" json:"os"`
	Platform        string `gorm:"type:varchar(100);not null" json:"platform"`
	PlatformFamily  string `gorm:"type:varchar(100);not null" json:"platform_family"`
	PlatformVersion string `gorm:"type:varchar(100);not null" json:"platform_version"`
	KernelVersion   string `gorm:"type:varchar(100);not null" json:"kernel_version"`
	KernelArch      string `gorm:"type:varchar(50);not null" json:"kernel_arch"`
	CPUCores        int    `gorm:"not null" json:"cpu_cores"`
	CPUModel        string `gorm:"type:text" json:"cpu_model"`
	TotalMemory     uint64 `gorm:"not null" json:"total_memory"`
	Uptime          uint64 `gorm:"not null" json:"uptime"`
	LastSeen        time.Time `gorm:"not null" json:"last_seen"`
}

// AlertRule 告警规则模型
type AlertRule struct {
	BaseModel
	Name        string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	MetricType  string    `gorm:"type:varchar(100);not null;index" json:"metric_type"` // cpu, memory, disk, network
	Operator    string    `gorm:"type:varchar(10);not null" json:"operator"`           // >, <, >=, <=, ==
	Threshold   float64   `gorm:"type:decimal(10,2);not null" json:"threshold"`
	Duration    int       `gorm:"not null" json:"duration"`     // 持续时间（秒）
	Severity    string    `gorm:"type:varchar(50);not null" json:"severity"` // info, warning, critical
	Enabled     bool      `gorm:"not null;default:true" json:"enabled"`
	Description string    `gorm:"type:text" json:"description"`
}

// Alert 告警记录模型
type Alert struct {
	BaseModel
	RuleID      uint      `gorm:"not null;index" json:"rule_id"`
	Rule        AlertRule `gorm:"foreignKey:RuleID" json:"rule"`
	Hostname    string    `gorm:"type:varchar(255);not null;index" json:"hostname"`
	MetricType  string    `gorm:"type:varchar(100);not null" json:"metric_type"`
	Value       float64   `gorm:"type:decimal(10,2);not null" json:"value"`
	Threshold   float64   `gorm:"type:decimal(10,2);not null" json:"threshold"`
	Severity    string    `gorm:"type:varchar(50);not null" json:"severity"`
	Message     string    `gorm:"type:text;not null" json:"message"`
	Status      string    `gorm:"type:varchar(50);not null;default:'active'" json:"status"` // active, resolved, suppressed
	StartTime   time.Time `gorm:"not null" json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Duration    *int      `json:"duration"` // 持续时间（秒）
}

// MonitoringConfig 监控配置模型
type MonitoringConfig struct {
	BaseModel
	Key         string `gorm:"type:varchar(255);not null;uniqueIndex" json:"key"`
	Value       string `gorm:"type:text;not null" json:"value"`
	Type        string `gorm:"type:varchar(50);not null" json:"type"`        // string, int, float, bool, json
	Category    string `gorm:"type:varchar(100);not null;index" json:"category"` // system, alert, dashboard
	Description string `gorm:"type:text" json:"description"`
	Editable    bool   `gorm:"not null;default:true" json:"editable"`
}

// Host 主机模型
type Host struct {
	BaseModel
	Hostname        string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"hostname"`
	DisplayName     string     `gorm:"type:varchar(255);not null" json:"display_name"`
	IPAddress       string     `gorm:"type:varchar(45);not null" json:"ip_address"` // IPv4 和 IPv6
	Environment     string     `gorm:"type:varchar(50);not null;index" json:"environment"` // prod, staging, dev, test
	Location        string     `gorm:"type:varchar(255)" json:"location"`
	Tags            string     `gorm:"type:text" json:"tags"` // JSON 格式存储标签
	Description     string     `gorm:"type:text" json:"description"`
	Status          string     `gorm:"type:varchar(20);not null;default:'unknown';index" json:"status"` // online, offline, maintenance, unknown
	MonitoringEnabled bool     `gorm:"not null;default:true" json:"monitoring_enabled"`
	LastSeen        *time.Time `json:"last_seen"`
	OS              string     `gorm:"type:varchar(100)" json:"os"`
	Platform        string     `gorm:"type:varchar(100)" json:"platform"`
	CPUCores        int        `json:"cpu_cores"`
	TotalMemory     uint64     `json:"total_memory"`
	Agent           bool       `gorm:"not null;default:false" json:"agent"` // 是否安装了监控代理
	
	// 关联关系
	Configs         []HostConfig `gorm:"foreignKey:HostID;constraint:OnDelete:CASCADE" json:"configs,omitempty"`
	Groups          []HostGroup  `gorm:"many2many:host_group_members;" json:"groups,omitempty"`
}

// HostConfig 主机配置模型
type HostConfig struct {
	BaseModel
	HostID      uint   `gorm:"not null;index" json:"host_id"`
	Key         string `gorm:"type:varchar(255);not null" json:"key"`
	Value       string `gorm:"type:text;not null" json:"value"`
	Type        string `gorm:"type:varchar(50);not null" json:"type"` // string, int, float, bool, json
	Category    string `gorm:"type:varchar(100);not null;index" json:"category"` // monitoring, alert, system
	Description string `gorm:"type:text" json:"description"`
	Editable    bool   `gorm:"not null;default:true" json:"editable"`
	
	// 关联关系
	Host        Host   `gorm:"foreignKey:HostID" json:"host,omitempty"`
}

// HostGroup 主机组模型
type HostGroup struct {
	BaseModel
	Name        string `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	DisplayName string `gorm:"type:varchar(255);not null" json:"display_name"`
	Description string `gorm:"type:text" json:"description"`
	Environment string `gorm:"type:varchar(50);index" json:"environment"`
	Tags        string `gorm:"type:text" json:"tags"` // JSON 格式存储标签
	Enabled     bool   `gorm:"not null;default:true" json:"enabled"`
	
	// 关联关系
	Hosts       []Host `gorm:"many2many:host_group_members;" json:"hosts,omitempty"`
}

// HostGroupMember 主机组成员关联表（用于多对多关系）
type HostGroupMember struct {
	HostID      uint      `gorm:"primaryKey" json:"host_id"`
	HostGroupID uint      `gorm:"primaryKey" json:"host_group_id"`
	JoinedAt    time.Time `gorm:"autoCreateTime" json:"joined_at"`
	Role        string    `gorm:"type:varchar(50);default:'member'" json:"role"` // member, admin
}

// TableName 设置表名
func (SystemMetrics) TableName() string {
	return "system_metrics"
}

func (SystemInfoDB) TableName() string {
	return "system_info"
}

func (AlertRule) TableName() string {
	return "alert_rules"
}

func (Alert) TableName() string {
	return "alerts"
}

func (MonitoringConfig) TableName() string {
	return "monitoring_configs"
}

func (Host) TableName() string {
	return "hosts"
}

func (HostConfig) TableName() string {
	return "host_configs"
}

func (HostGroup) TableName() string {
	return "host_groups"
}

func (HostGroupMember) TableName() string {
	return "host_group_members"
}